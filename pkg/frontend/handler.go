package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const home = `<html><head><title>Pleiades</title></head><body>
<h1>Pleiades</h1>
<p>Welcome to Pleiades</p>
</body></html>`

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, home)
}

func (f *Frontend) websocketHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (f *Frontend) statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp := &Counters{}
	counters, err := f.getAllCounters(ctx)
	if err != nil {
		logger.Errorf("Error retrieving Redis stats: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.Counters = counters
	b, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Error marshalling stats respone: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func (f *Frontend) getKeys(ctx context.Context) ([]string, error) {
	keys, err := f.r.Keys(ctx, "pleiades*").Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (f *Frontend) getAllCounters(ctx context.Context) ([]Counter, error) {
	keys, err := f.getKeys(ctx)
	if err != nil {
		return nil, err
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
			Name:        k,
			Description: "",
			Value:       parsedVal,
		}
	}
	return out, nil
}

func (f *Frontend) getCounter(ctx context.Context, name string) (Counter, error) {
	c := Counter{}
	return c, nil
}
