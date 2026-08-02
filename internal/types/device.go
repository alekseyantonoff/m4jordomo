// Канонические типы устройств и реестр типов событий
package types

import "time"

// Типы событий
const (
	// Поток А: команды: ядро → устройство
	EventCommandDeviceSet    = "command.device.set"
	EventCommandDeviceGetAll = "command.device.get_all"

	// Поток Б: данные: устройство → ядро
	EventDeviceStateChanged  = "device.state.changed"
	EventDeviceStateResponse = "device.state.response"
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

// EventDeadLetter — событие не доставлено после всех попыток
const EventDeadLetter = "system.dead_letter"

// DeadLetter — запись для «мёртвой очереди»: что не доставилось и почему
type DeadLetter struct {
	Event    Event  // исходное событие, которое не удалось доставить
	Reason   string // причина провала
	Attempts int    // сколько раз пытались доставить
}
