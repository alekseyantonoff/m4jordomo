// Плагин персистентности: сохраняет недоставленные события (DLQ) в БД
package persistence

import (
	"encoding/json"
	"log"

	"m4jordomo/internal/bus"
	"m4jordomo/internal/storage"
	"m4jordomo/internal/types"
)

// PersistencePlugin — слушает system.dead_letter и пишет в таблицу dead_letters
type PersistencePlugin struct {
	storage *storage.Storage
}

// New — создаёт плагин
func New(st *storage.Storage) *PersistencePlugin {
	return &PersistencePlugin{storage: st}
}

// Name — имя плагина
func (p *PersistencePlugin) Name() string {
	return "persistence"
}

// Init — подписывается на события DLQ
func (p *PersistencePlugin) Init(b *bus.Bus) error {
	log.Println("[Persistence] Подписываюсь на system.dead_letter...")

	b.Subscribe(types.EventDeadLetter, func(e types.Event) {
		p.handleDeadLetter(e)
	})

	log.Println("[Persistence] Подписки выполнены.")
	return nil
}

// handleDeadLetter — сериализует недоставленное событие и сохраняет в БД
func (p *PersistencePlugin) handleDeadLetter(e types.Event) {
	dl, ok := e.Payload.(types.DeadLetter)
	if !ok {
		log.Printf("[Persistence] ❌ Ошибка: неверный формат payload: %v", e.Payload)
		return
	}

	payloadJSON, err := json.Marshal(dl.Event)
	if err != nil {
		log.Printf("[Persistence] ❌ Ошибка сериализации события: %v", err)
		return
	}

	rec := storage.DeadLetterRecord{
		EventType: dl.Event.Type,
		Priority:  int(dl.Event.Priority),
		Payload:   string(payloadJSON),
		Reason:    dl.Reason,
		Attempts:  dl.Attempts,
	}

	if err := p.storage.SaveDeadLetter(rec); err != nil {
		log.Printf("[Persistence] ❌ Ошибка сохранения в DLQ: %v", err)
		return
	}
	log.Printf("[Persistence] 💾 Записано в DLQ: %s (попыток: %d)", dl.Event.Type, dl.Attempts)
}
