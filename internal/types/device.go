// Канонические типы устройств и реестр типов событий
package types

import "time"

// Типы событий
const (
	// Поток А: команды: ядро → устройство
	EventCommandDeviceSet     = "command.device.set"
	EventCommandDeviceGetAll  = "command.device.get_all"
	EventCommandDeviceBreakIt = "command.device.break_it" // тестовая: публикуется без подписчиков → уходит в DLQ

	// Поток Б: данные: устройство → ядро
	EventDeviceStateChanged  = "device.state.changed"
	EventDeviceStateResponse = "device.state.response"

	// DLQ: недоставленные события
	EventDeadLetter              = "system.dead_letter"          // событие-запись в DLQ
	EventCommandDeadLetterGetAll = "command.dead_letter.get_all" // команда: показать список DLQ
	EventCommandDeadLetterReplay = "command.dead_letter.replay"  // команда «повторить недоставленные события»
	EventDeadLetterResponse      = "dead_letter.response"        // ответ: список DLQ
)

// DeviceCommand — команда управления устройством (поток А).
// Будет лежать в Event.Payload при публикации command.device.set.
type DeviceCommand struct {
	Name   string // имя устройства: "ac", "light"
	Status bool   // true = включить/открыть, false = выключить/закрыть
}

// DeviceState — состояние, которое устройство сообщило само (поток Б).
type DeviceState struct {
	Name      string
	Status    bool
	UpdatedAt time.Time // когда устройство прислало
}

// SensorReading — показание датчика (поток Б).
type SensorReading struct {
	Name   string  // имя датчика: "temp_sensor"
	Metric string  // что измеряем: "temperature"
	Value  float64 // значение: 22.5
	Unit   string  // единица: "C"
}

// DeviceStateList — снимок всех состояний (ответ на command.device.get_all)
type DeviceStateList struct {
	States map[string]bool // name -> статус
}

// DeadLetter — запись для «мёртвой очереди»: что не доставилось и почему
type DeadLetter struct {
	Event    Event  // исходное событие, которое не удалось доставить
	Reason   string // причина провала
	Attempts int    // сколько раз пытались доставить
}

// DeadLetterRecord — строка таблицы dead_letters (ответ на command.dead_letter.get_all)
type DeadLetterRecord struct {
	ID        int64  // автоинкремент из БД
	EventType string // тип события, которое не доставилось
	Priority  int    // приоритет (число)
	Payload   string // сериализованный payload в JSON
	Reason    string // причина провала
	Attempts  int    // сколько раз пытались
}

// DeadLetterList — снимок списка DLQ (ответ на command.dead_letter.get_all)
type DeadLetterList struct {
	Records []DeadLetterRecord // записи из таблицы dead_letters
}
