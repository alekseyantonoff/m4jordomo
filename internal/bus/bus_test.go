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

func TestAllHandlersAreCalled(t *testing.T) {
	count := 0
	b := New()

	b.Subscribe("test.fanout", func(_ types.Event) {
		count++
	})

	b.Subscribe("test.fanout", func(_ types.Event) {
		count++
	})

	b.Publish(types.Event{
		Type:     "test.fanout",
		Priority: types.Medium,
	})

	const want = 2
	if count != want {
		t.Errorf("вызвано обработчиков: got=%d, want=%d", count, want)
	}
}
