package file

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gargath/pleiades/pkg/aggregator"
	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/util"
)

const moduleName = "file-agg"

var (
	// ErrNoSrc is returned when an Aggregator is created without a source directory
	ErrNoSrc = fmt.Errorf("No source directory provided")

	procTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pleiades_aggregator_file_process_duration_milliseconds",
			Help:    "Time taken to process files",
			Buckets: []float64{5, 10, 100, 500},
		},
	)

	logger = log.MustGetLogger(moduleName)
	wg     sync.WaitGroup
)

// NewAggregator returns a Aggregator initialized with the source path provided
func NewAggregator(redisOpts *util.RedisOpts, opts *Opts) (*Aggregator, error) {
	a := &Aggregator{}
	src := opts.Source
	if src == "" {
		return nil, ErrNoSrc
	}
	o, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory %s does not exist", src)
	} else if o.Mode().IsRegular() {
		logger.Errorf("source path %s exists and is file", src)
		return nil, fmt.Errorf("source path %s exists as file", src)
	}

	r, err := util.NewValidatedRedisClient(redisOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", redisOpts.RedisAddr, err)
	}

	a.r = r
	a.File = opts
	a.Redis = redisOpts
	a.stop = make(chan (bool))

	return a, nil
}

// Start starts up the aggregation server
func (a *Aggregator) Start() error {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.stop:
				{
					return
				}
			default:
				err := a.run()
				if err != nil {
					logger.Errorf("Aggregator exited with error: %v", err)
				}
			}
		}
	}()

	if !util.IsTTY() {
		logger.Info("Terminal is not a TTY, not displaying progress indicator")
	} else {
		a.spinner = util.NewSpinner("Processing... ")
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-a.stop:
					return
				default:
					a.spinner.Tick()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

// Stop shuts down the aggregation server
func (a *Aggregator) Stop() {
	close(a.stop)
	wg.Wait()
}

func (a *Aggregator) run() error {
	for {
		start := time.Now()
		logger.Debugf("Reading directory listing for %s", a.File.Source)
		files, err := ioutil.ReadDir(a.File.Source)
		logger.Debugf("Listing directory took %s", time.Since(start))
		if err != nil {
			return err
		}
		if len(files) == 0 {
			select {
			case <-a.stop:
				return nil
			default:
				logger.Info("No files in source directory - will try again in 5 seconds")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				a.r.Ping(ctx)
				time.Sleep(5 * time.Second)
			}
		} else {
			for _, f := range files {
				select {
				case <-a.stop:
					return nil
				default:
					err := a.processFile(a.File.Source + "/" + f.Name())
					if err != nil {
						logger.Errorf("Error processing file %s: %v", f.Name(), err)
					}
				}
			}
		}
	}
}

func (a *Aggregator) processFile(filename string) error {
	defer func(start time.Time) {
		procTime.Observe(float64(time.Since(start).Milliseconds()))
	}(time.Now())

	fh, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("unreadable file %s: %v", filename, err)
	}
	scanner := bufio.NewScanner(fh)
	if !scanner.Scan() {
		return fmt.Errorf("premature end of file while reading %s", filename)
	}
	msgID := scanner.Text()
	if !scanner.Scan() {
		return fmt.Errorf("premature end of file while reading %s", filename)
	}
	eventData := scanner.Bytes()
	if scerr := scanner.Err(); scerr != nil {
		return fmt.Errorf("failed to read data from file %s: %v", filename, scerr)
	}
	fh.Close()

	counters, lendiff, err := aggregator.CountersFromEventData(eventData)
	aggregator.RecordLag(msgID)
	if err != nil {
		return fmt.Errorf("error processing file %s: %v", filename, err)
	}
	// TODO: this is duplicatede between the two aggregators. Should refactor.

	eventTimestamp, err := aggregator.ParseTimestamp(string(msgID))
	if err != nil {
		return fmt.Errorf("failed to parse timestamp from message: %s: %v", string(msgID), err)
	}
	var julianDay int64 = eventTimestamp / 86400000
	julianPrefix := fmt.Sprintf("day_%d_", julianDay)

	for _, counter := range counters {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := a.r.Incr(ctx, counter).Err()
		if err != nil {
			return fmt.Errorf("failed to increment Redis counter %s: %v", counter, err)
		}
		err = a.r.Incr(ctx, julianPrefix+counter).Err()
		if err != nil {
			return fmt.Errorf("failed to increment Redis counter %s: %v", julianPrefix+counter, err)
		}
	}
	// TODO: remove that duplication below once the return from CountersFromEventData() is less stupid
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = a.r.IncrBy(ctx, "pleiades_growth", lendiff).Err()
	if err != nil {
		return fmt.Errorf("failed to increment Redis growth counter: %v", err)
	}
	err = a.r.IncrBy(ctx, julianPrefix+"pleiades_growth", lendiff).Err()
	if err != nil {
		return fmt.Errorf("failed to increment historic Redis growth counter: %v", err)
	}

	err = os.Remove(filename)
	if err != nil {
		return fmt.Errorf("failed to delete source file %s: %v", filename, err)
	}
	return nil
}
