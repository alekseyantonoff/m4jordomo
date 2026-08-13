package bus

import (
	"m4jordomo/internal/types"
	"testing"
)

func TestSubscribeAndPublish(t *testing.T) {
	var received bool
	b := New()
	b.Subscribe("test.basic", func(_ types.Event) {
		received = true
	})

	b.Publish(types.Event{
		Type:     "test.basic",
		Priority: types.Medium,
	})

	if !received {
		t.Errorf("обработчик не вызван: received=%v", received)
	}
}
