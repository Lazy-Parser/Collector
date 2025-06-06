package mexc

import (
	"context"
	"encoding/json"
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/ui"
)

type Mexc struct {
	Pool *Pool
}

func (m *Mexc) Name() string {
	return "MEXC"
}

func (m *Mexc) Connect() error {

	return nil
}

func (m *Mexc) Subscribe(ctx context.Context, pairs []database.Pair) error {
	for _, pair := range pairs {
		if err := m.Pool.Subscribe(ctx, pair.BaseToken.Name); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mexc) Run(ctx context.Context, consumerChan chan core.MexcResponse) {
	for {
		select {
		case <-ctx.Done():
			m.Pool.UnsubscribeAll()
			return

		case msgBytes := <-m.Pool.Listen():
			body, err := handleMsg(msgBytes)
			if err != nil {
				//ui.GetUI().LogsView(fmt.Sprintf("[MEXC] error handleMsg: %v", err))
				ui.GetUI().LogsView(string(msgBytes), "warning")
				continue
			}

			consumerChan <- body
		}
	}
}

// methods
func handleMsg(msg []byte) (core.MexcResponse, error) {
	var res core.MexcResponse

	if err := json.Unmarshal([]byte(msg), &res); err != nil {
		return res, err
	}

	return res, nil
}