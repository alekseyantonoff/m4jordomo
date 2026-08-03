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

	b.Subscribe(types.EventCommandDeadLetterGetAll, func(_ types.Event) {
		p.handleGetAllDeadLetters(b)
	})

	b.Subscribe(types.EventCommandDeadLetterReplay, func(_ types.Event) {
		p.handleReplayDeadLetters(b)
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

	rec := types.DeadLetterRecord{
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

// handleGetAllDeadLetters — получает все записи из таблицы dead_letters и публикует ответ
func (p *PersistencePlugin) handleGetAllDeadLetters(bus *bus.Bus) {
	records, err := p.storage.GetDeadLetters()
	if err != nil {
		log.Printf("[Persistence] ❌ Ошибка получения данных из dead_letters: %v", err)
		return
	}

	bus.Publish(types.Event{
		Type:     types.EventDeadLetterResponse,
		Priority: types.Medium,
		Payload: types.DeadLetterList{
			Records: records,
		},
	})
}

// handleReplayDeadLetters — пробует доставить записи из DLQ повторно,
// при успехе удаляет их из очереди
func (p *PersistencePlugin) handleReplayDeadLetters(bus *bus.Bus) {
	records, err := p.storage.GetDeadLetters()
	if err != nil {
		log.Printf("[Persistence] ❌ Ошибка получения данных из dead_letters: %v", err)
		return
	}

	log.Println("[Persistence] 🔁 Начинаю replay недоставленных событий...")
	for _, rec := range records {
		event, err := restoreEvent(rec)
		if err != nil {
			log.Printf("[Persistence] ❌ Не удалось восстановить событие #%d: %v", rec.ID, err)
			continue
		}

		err = bus.PublishOnce(event)
		if err != nil {
			log.Printf("[Persistence] ⏭️ #%d не доставлено повторно: %v", rec.ID, err)
			continue
		}

		if err := p.storage.DeleteDeadLetter(rec.ID); err != nil {
			log.Printf("[Persistence] ❌ Не удалось удалить запись #%d: %v", rec.ID, err)
			continue
		}
		log.Printf("[Persistence] ✅ #%d доставлено повторно и удалено из DLQ", rec.ID)
	}
	log.Println("[Persistence] 🔁 Реплей завершён.")
}

// restoreEvent - восстанавливает типизированное событие из JSON-записи DLQ
func restoreEvent(rec types.DeadLetterRecord) (types.Event, error) {
	var e types.Event
	if err := json.Unmarshal([]byte(rec.Payload), &e); err != nil {
		return e, err
	}

	// e.Payload после Unmarshal — это map[string]interface{}.
	// Перегоняем его в типизированную структуру по типу события.
	switch e.Type {
	case types.EventCommandDeviceSet, types.EventCommandDeviceBreakIt:
		var cmd types.DeviceCommand
		data, err := json.Marshal(e.Payload)
		if err != nil {
			return e, err
		}
		if err := json.Unmarshal(data, &cmd); err != nil {
			return e, err
		}
		e.Payload = cmd
	}

	e.RetryCount = 0
	return e, nil
}
