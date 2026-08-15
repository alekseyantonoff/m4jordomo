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

func TestSendToDeadLetter(t *testing.T) {
	b := New()

	// prepare
	var dl types.DeadLetter
	gotDL := false
	b.Subscribe(types.EventDeadLetter, func(e types.Event) {
		v, ok := e.Payload.(types.DeadLetter)
		if !ok {
			t.Fatalf("Payload не DeadLetter: %T", e.Payload)
		}
		dl = v
		gotDL = true
	})

	// act
	b.sendToDeadLetter(types.Event{
		Type:     "command.device.break_it",
		Priority: types.High,
	})

	// assert
	if !gotDL {
		t.Fatal("событие не пришло в system.dead_letter")
	}
	if dl.Event.Type != "command.device.break_it" {
		t.Errorf("Event.Type: got=%q, want=%q", dl.Event.Type, "command.device.break_it")
	}
	if dl.Reason == "" {
		t.Error("Reason пустой")
	}
}

func TestHighWithNoSubscribersGoesToDLQ(t *testing.T) {
	b := New()

	// prepare: подписка на DLQ ДО публикации
	var dl types.DeadLetter
	gotDL := false
	b.Subscribe(types.EventDeadLetter, func(e types.Event) {
		v, ok := e.Payload.(types.DeadLetter)
		if !ok {
			t.Fatalf("Payload не DeadLetter: %T", e.Payload)
		}
		dl = v
		gotDL = true
	})

	// act: High без подписчиков -> 3 попытки по 1с, потом DLQ (~2с)
	b.Publish(types.Event{
		Type:     "command.device.break_it",
		Priority: types.High,
	})

	// assert
	if !gotDL {
		t.Fatal("High-событие не попало в system.dead_letter")
	}
	if dl.Event.Type != "command.device.break_it" {
		t.Errorf("Event.Type: got=%q, want=%q", dl.Event.Type, "command.device.break_it")
	}
	if dl.Reason == "" {
		t.Error("Reason пустой")
	}
	if dl.Attempts != 3 {
		t.Errorf("Attempts: got=%d, want=%d", dl.Attempts, 3)
	}
}

func TestPanicInHandlerDoesNotStopNext(t *testing.T) {
	b := New()

}
