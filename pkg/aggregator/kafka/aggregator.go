package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gargath/pleiades/pkg/aggregator"
	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/spinner"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/segmentio/kafka-go"
)

const moduleName = "kafka-agg"

var (
	wg sync.WaitGroup

	logger      = log.MustGetLogger(moduleName)
	kafkaLogger = log.MustGetLogger("kafka-client")

	// ErrNoSrc is returned when an Aggregator is created without a kafka source
	ErrNoSrc = fmt.Errorf("No source kafka details provided")

	procTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pleiades_aggregator_kafka_process_duration_milliseconds",
			Help:    "Time taken to process event from kafka",
			Buckets: []float64{5, 10, 100, 500},
		},
	)

	msgTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pleiades_aggregator_event_count_total",
			Help: "Number of events processed",
		},
	)
)

// NewAggregator returns a Aggregator initialized with the kafka details provided
func NewAggregator(redisOpts *aggregator.RedisOpts, opts *Opts) (*Aggregator, error) {
	a := &Aggregator{}
	broker := opts.Broker
	topic := opts.Topic
	if (broker == "") || (topic == "") {
		return nil, ErrNoSrc
	}

	k := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		GroupID:        "pleiades-aggregator-group",
		Topic:          topic,
		CommitInterval: time.Second,
		ErrorLogger:    &crudErrorLogger{},
		Logger:         newCrudLogger(),
	})

	r := redis.NewClient(&redis.Options{
		Addr: redisOpts.RedisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := r.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", redisOpts.RedisAddr, err)
	}
	logger.Debugf("Connected to Redis: %v", pong)

	a.r = r
	a.Kafka = opts
	a.Redis = redisOpts
	a.k = k
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

	if !spinner.IsTTY() {
		logger.Info("Terminal is not a TTY, not displaying progress indicator")
	} else {
		a.spinner = spinner.NewSpinner("Processing... ")
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
		select {
		case <-a.stop:
			return nil
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := a.k.ReadMessage(ctx)
			if ctx.Err() == context.DeadlineExceeded {
				cancel()
				logger.Info("No new messages on topic for 5 seconds. Will try again")
				continue
			}
			defer cancel()
			if err != nil {
				logger.Errorf("Error reading message from kafka: %v", err)
			}
			var pErr error
			pErr = a.processEvent(msg.Key, msg.Value)
			return pErr
		}
	}
}

func (a *Aggregator) processEvent(id []byte, data []byte) error {
	defer func(start time.Time) {
		procTime.Observe(float64(time.Since(start).Milliseconds()))
	}(time.Now())

	counters, lendiff, err := aggregator.CountersFromEventData(data)
	aggregator.RecordLag(string(id))
	if err != nil {
		return fmt.Errorf("error processing event: %s, %v", string(data), err)
	}
	for _, counter := range counters {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := a.r.Incr(ctx, counter).Err()
		if err != nil {
			return fmt.Errorf("failed to increment Redis counter: %v", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = a.r.IncrBy(ctx, "pleiades_growth", lendiff).Err()
	if err != nil {
		return fmt.Errorf("failed to increment Redis growth counter: %v", err)
	}

	msgTotal.Inc()
	return nil
}
