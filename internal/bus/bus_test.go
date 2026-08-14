package bus

import (
	"errors"
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
		t.Errorf("Обработчик не вызван: received=%v", received)
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
		t.Errorf("Вызвано обработчиков: got=%d, want=%d", count, want)
	}
}

func TestPublishOnceUnknownTypeReturnsError(t *testing.T) {
	const noSuchType = "no.such.type"
	b := New()
	err := b.publishOnce(types.Event{Type: noSuchType, Priority: types.Medium})

	if err == nil || !errors.Is(err, ErrNoSubscribers) {
		t.Errorf("publishOnce(%s): want %v, got %v", noSuchType, ErrNoSubscribers, err)
	}
}
