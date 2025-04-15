package aggregator

import (
	"log"
	"strings"
	"sync"
)

var (
	once   sync.Once
	joiner *Joiner
	bus    chan AggregatorStruct
)

func InitJoiner() {
	once.Do(func() {
		joiner = &Joiner{
			cache: make(map[string]map[string]AggregatorStruct),
		}
		bus = make(chan AggregatorStruct, 5000) // буферизированный канал
		go joiner.run()
	})
}

func GetJoiner() *Joiner {
	return joiner
}

func (j *Joiner) run() {
	log.Println("🧠 Aggregator started...")

	for data := range bus {
		j.Update(data)
	}
}

func (j *Joiner) Push(data AggregatorStruct) {
	bus <- data
}

// Обновление внутреннего кэша и объединение
func (j *Joiner) Update(data AggregatorStruct) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.cache[data.Symbol]; !ok {
		j.cache[data.Symbol] = make(map[string]AggregatorStruct)
	}
	j.cache[data.Symbol][data.Type] = data

	// Пример: логика сравнения цен с разных бирж
	futures := j.cache[data.Symbol]["FUTURES"]
	spot := j.cache[data.Symbol]["SPOT"]

	// Если все три есть, покажем расхождение
	if futures.Price != "" && spot.Price != "" {
		log.Printf("🔁 %s: FUTURES %s | SPOT %s",
			data.Symbol, futures.Price, spot.Price)
	}
}

func NormalizeSymbol(symbol string) string {
	newString := strings.ReplaceAll(symbol, "_", "/")

	if newString == symbol {
		newString = strings.ReplaceAll(symbol, "USDT", "/USDT")
	}

	return newString
}

