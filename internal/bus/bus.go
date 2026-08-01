// Шина событий
package bus

import (
	"log"
	"m4jordomo/internal/types"
	"sync"
	"time"
)

// Bus — шина событий
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(types.Event)
}

// New — создает новую шину
func New() *Bus {
	return &Bus{
		subscribers: make(map[string][]func(types.Event)),
	}
}

// Subscribe — подписывает функцию на событие
func (b *Bus) Subscribe(eventType string, fn func(types.Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], fn)
	log.Printf("[Bus] Подписка: %s -> %d обработчиков", eventType, len(b.subscribers[eventType]))
}

// Publish — публикует событие в шину
func (b *Bus) Publish(event types.Event) {
	event.Timestamp = time.Now()
	log.Printf("[Bus] 📤 Публикация: %s (приоритет: %s)", event.Type, event.Priority.String())

	switch event.Priority {
	case types.Critical:
		b.publishWithRetry(event, 5, 500*time.Millisecond)
	case types.High:
		b.publishWithRetry(event, 3, 1*time.Second)
	case types.Medium:
		b.publishWithRetry(event, 1, 2*time.Second)
	case types.Low:
		b.publishOnce(event)
	}
}

// publishOnce — пытается доставить событие один раз
func (b *Bus) publishOnce(event types.Event) error {
	b.mu.RLock()
	handlers, exists := b.subscribers[event.Type]
	b.mu.RUnlock()

	if !exists {
		log.Printf("[Bus] Нет подписчиков на событие: %s", event.Type)
		return nil
	}

	for _, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Bus] 🔥 Паника в обработчике: %v", r)
				}
			}()
			handler(event)
		}()
	}
	return nil
}

// publishWithRetry — пытается доставить событие несколько раз
func (b *Bus) publishWithRetry(event types.Event, maxRetries int, delay time.Duration) {
	for i := 0; i < maxRetries; i++ {
		err := b.publishOnce(event)
		if err == nil {
			log.Printf("[Bus] ✅ Событие %s доставлено (попытка %d)", event.Type, i+1)
			return
		}
		log.Printf("[Bus] ⚠️ Ошибка доставки %s (попытка %d/%d): %v", event.Type, i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}

	log.Printf("[Bus] 🔴 КРИТИЧЕСКАЯ ОШИБКА: Событие %s потеряно после %d попыток", event.Type, maxRetries)

	if event.Priority == types.Critical {
		b.publishOnce(types.Event{
			Type:     "system.emergency.trigger",
			Priority: types.Critical,
			Payload: map[string]interface{}{
				"original_event": event.Type,
				"reason":         "Превышено количество попыток доставки",
			},
		})
	}
}

// Debug — печатает всех подписчиков (для отладки)
func (b *Bus) Debug() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	log.Println("=== Текущие подписки ===")
	for eventType, handlers := range b.subscribers {
		log.Printf("  %s: %d обработчиков", eventType, len(handlers))
	}
}
