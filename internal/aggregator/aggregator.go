package aggregator

import (
	// "encoding/json"
	"fmt"
	"reflect"

	// "log"
	"sync"
	"time"

	m "github.com/Lazy-Parser/Collector/internal/models"
	"github.com/nats-io/nats.go"
)

var (
	joiner *Joiner
	once   sync.Once
)

type Joiner struct {
	nats   *nats.Conn
	buffer map[string]*m.Coin
	mu     sync.Mutex
}

// InitJoiner инициализируется один раз
func InitJoiner() {
	once.Do(func() {
		joiner = &Joiner{
			buffer: make(map[string]*m.Coin),
		}
	})
}

// Глобальный доступ
func GetJoiner() *Joiner {
	return joiner
}

func (j *Joiner) UpdateMexc(pair string, info *m.MexcInfo) {
	j.mu.Lock()
	defer j.mu.Unlock()

	coin, exists := j.buffer[pair]
	if !exists {
		coin = &m.Coin{Pair: pair}
		j.buffer[pair] = coin
	}

	coin.Mexc = info
	coin.MexcReady = true

	j.checkAndSend(coin)
}

func (j *Joiner) UpdateDex(pair string, info *m.DexInfo) {
	j.mu.Lock()
	defer j.mu.Unlock()

	coin, exists := j.buffer[pair]
	if !exists {
		coin = &m.Coin{Pair: pair}
		j.buffer[pair] = coin
	}

	coin.Dex = info
	coin.DexReady = true

	j.checkAndSend(coin)
}

func (j *Joiner) checkAndSend(c *m.Coin) {
	// if !c.DexReady || !c.MexcReady {
	// 	return
	// }

	c.Timestamp = time.Now()
	print(c)
	// payload, _ := json.Marshal(c)

	// err := j.nats.Publish("coins", payload)
	// if err != nil {
	// 	log.Fatalf("publish message coin: %w", err)
	// }

	delete(j.buffer, c.Pair)
}

// delete
func print(c *m.Coin) {
	value := reflect.ValueOf(c)
	key := reflect.TypeOf(c)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		key = key.Elem()
	}

	fmt.Println("\n");
	for i := range value.NumField() {
		fmt.Println("%s: %s", key.Field(i).Name, value.Field(i).Interface())
	}
}