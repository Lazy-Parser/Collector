package metrics

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	activeConnectionsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_active_connections_total",
		Help: "Current number of WebSocket connections",
	})

	activeConnectionsMexc = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_active_connections_Mexc",
		Help: "Current number of WebSocket connections on Mexc Client",
	})

	activeCoinsMexc = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_active_coins_mexc",
		Help: "Total number of coins to listen on Mexc",
	})

	messagesReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_received_total",
		Help: "Total number of messages received from ALL exchanges",
	})

	messagesReceivedMexc = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_received_Mexc",
		Help: "Total number of messages received from Mexc",
	})

	messagesSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_sent_total",
		Help: "Total number of messages sent to APIs",
	})

	messagesSentMexc = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_sent_Mexc",
		Help: "Total number of messages sent to Mexc API",
	})

	wsErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_errors_total",
		Help: "Total number of WebSocket errors",
	})

	// TODO: delete
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{Name: "myapp_processed_ops_total", Help: "The total number of processed events"})
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func StartMetrics() error {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Start Prometheus on http://localhost:2112/metrics")
	if err := http.ListenAndServe(":2112", nil); err != nil {
		return err
	}

	return nil
}
