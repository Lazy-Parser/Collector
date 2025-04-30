// Package aggregator implements the core logic for aggregating real-time
// market data from various sources such as MEXC Spot, Futures, and others.
//
// It handles normalization, caching, deduplication, and comparison of prices.
// Once data from different sources for the same symbol is available, it
// publishes a unified message (e.g., spread) for external use.
//
// This package is designed to be thread-safe and run concurrently in the background.
// It uses channels and sync primitives to ensure safe access to shared data.

package aggregator

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	d "github.com/Lazy-Parser/Collector/internal/domain"
	p "github.com/Lazy-Parser/Collector/internal/impl/publisher"
)

var (
	once        sync.Once
	joiner      *Joiner
	bus         chan d.AggregatorPayload
	stream      chan d.AggregatorPayload
	isStreaming chan bool // when some collector stop working, set to false
)

// InitJoiner initializes the singleton Joiner and NATS connection structures.
func InitJoiner() {
	once.Do(func() {
		joiner = &Joiner{
			cache: make(map[string]map[string]d.AggregatorPayload),
		}
		bus = make(chan d.AggregatorPayload, 5000)
		stream = make(chan d.AggregatorPayload, 5000)
		isStreaming = make(chan bool, 1)
		isStreaming <- true // by default true

		// init nats connect to publish joiners
		p.InitPublisher()
		// defer p.GetPublisher().Close()

		go joiner.run()
	})
}

func GetJoiner() *Joiner {
	return joiner
}

// run listens to the bus channel for incoming AggregatorStruct data.
func (j *Joiner) run() {
	fmt.Println("🧠 Aggregator started...")

	for data := range bus {
		j.Update(data)
	}
}

// add data from param to the bus. Data will be added to cache . OLD: AggregatorStruct
func (j *Joiner) Push(data d.AggregatorPayload) {
	bus <- data
	stream <- data
}

// set flag if work or not
func (j *Joiner) SetState(value bool) {
	isStreaming <- value
}

// return all data, that collectors are send to aggregator. It usefull for a not main logic (generator)
func (j *Joiner) Stream() <-chan d.AggregatorPayload {
	return stream
}

func (j *Joiner) ListenState() <-chan bool {
	return isStreaming
}

// Update cache and aggregate all data
func (j *Joiner) Update(data d.AggregatorPayload) {
	j.mu.Lock()
	// defer j.mu.Unlock()

	// if _, ok := j.cache[data.Symbol]; !ok {
	// 	j.cache[data.Symbol] = make(map[string]d.AggregatorPayload)
	// }
	// //		"MEXC"
	// j.cache[data.Symbol][data.Type] = data

	// futures := j.cache[data.Symbol]["FUTURES"].Futures
	// spot := j.cache[data.Symbol]["SPOT"].Spot

	// // send via nats
	// if futures.LastPrice != 0 && spot.Price != 0 {
	// 	payload := &p.Message{
	// 		Symbol:    data.Symbol,
	// 		Timestamp: data.Timestamp.UnixMilli(),
	// 		Spot:      spot,
	// 		Futures:   futures,
	// 	}

	// 	err := p.GetPublisher().Publish("mexc.spread", *payload)
	// 	if err != nil {
	// 		fmt.Errorf("Send message erorr: %w", err)
	// 	}

	// fmt.Printf("🔁 %s: FUTURES %s | SPOT %s",
	// 	data.Symbol, futures.Price, spot.Price)
	// }
}

func Print(obj interface{}, indent string) {
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			fmt.Printf("%s%s:\n", indent, fieldType.Name)
			print(field.Interface(), indent+"  ")
		} else {
			fmt.Printf("%s%s: %v\n", indent, fieldType.Name, field.Interface())
		}
	}
}

// return the latest time
func max(t1 time.Time, t2 time.Time) time.Time {
	if t1.Compare(t2) == -1 { // t1 < t2
		return t2
	}

	return t1
}
