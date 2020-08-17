package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	counterDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pleiades_web_counter_marshal_duration_seconds",
		Help: "Time taken to generate the stats json",
	}, []string{"operation"})
)

func (f *Frontend) websocketHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (f *Frontend) daysHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") //remove later

	days, err := f.getDays(ctx)
	if err != nil {
		logger.Errorf("Error retrieving available days from Redis keys: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if days == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, err := json.Marshal(days)
	if err != nil {
		logger.Errorf("Error marshalling stats respone: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func (f *Frontend) statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") //remove later

	julianDay := time.Now().Unix() / 86400

	counters, err := f.getAllCounters(ctx, julianDay)
	if err != nil {
		logger.Errorf("Error retrieving Redis stats: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if counters == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &Counters{
		Since:    julianDay * 86400,
		Counters: counters,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Error marshalling stats respone: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func (f *Frontend) statsForDayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	julianDay, err := strconv.ParseInt(vars["day"], 10, 64)
	if err != nil {
		logger.Infof("Rejecting invalid day format %s: %v", vars["day"], err)
		w.WriteHeader(http.StatusBadRequest) //TODO: Add an error response that is useful
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") //remove later

	counters, err := f.getAllCounters(ctx, julianDay)
	if err != nil {
		logger.Errorf("Error retrieving Redis stats: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if counters == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &Counters{
		Since:    julianDay * 86400,
		Counters: counters,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Error marshalling stats respone: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func (f *Frontend) getKeys(ctx context.Context, day int64) ([]string, error) {
	prefix := fmt.Sprintf("day_%d_pleiades*", day)
	logger.Debugf("getting counters for prefix %s", prefix)
	keys, err := f.r.Keys(ctx, prefix).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (f *Frontend) getAllCounters(ctx context.Context, julianDay int64) ([]Counter, error) {
	timer := prometheus.NewTimer(counterDuration.WithLabelValues("get_counters"))

	prefix := fmt.Sprintf("day_%d_", julianDay)
	keys, err := f.getKeys(ctx, julianDay)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	out := make([]Counter, len(keys))
	result, error := f.r.MGet(ctx, keys...).Result()
	if error != nil {
		return nil, error
	}
	for i, k := range keys {
		var val string
		var ok bool
		if val, ok = result[i].(string); !ok {
			return nil, fmt.Errorf("invalid non-string type in Redis counter %s: %v ", k, result[i])
		}
		parsedVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("redis value not parsable as number: %s - %v", val, err)
		}
		out[i] = Counter{
			Name:        strings.SplitAfter(k, prefix)[1],
			Description: "",
			Value:       parsedVal,
		}
	}
	timer.ObserveDuration()
	return out, nil
}

func (f *Frontend) getCounter(ctx context.Context, name string) (Counter, error) {
	c := Counter{}
	return c, nil
}

func (f *Frontend) getDays(ctx context.Context) ([]Day, error) {
	timer := prometheus.NewTimer(counterDuration.WithLabelValues("get_days"))

	keys, err := f.r.Keys(ctx, "day_*").Result()
	if err != nil {
		return nil, err
	}

	uniqueDays := make(map[string]bool)
	for _, v := range keys {
		d := strings.Split(v, "_")[1]
		dNum, _ := strconv.Atoi(d)
		if dNum > 18488 {
			uniqueDays[d] = true
		}
	}
	d := []string{}
	for k := range uniqueDays {
		d = append(d, k)
	}
	sort.Strings(d)
	out := make([]Day, len(d))
	for i, v := range d {
		out[i] = Day(v)
	}
	timer.ObserveDuration()
	return out, nil
}
