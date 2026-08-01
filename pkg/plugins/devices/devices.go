// Плагин управления устройствами
package devices

import (
	"log"
	"m4jordomo/internal/bus"
	"m4jordomo/internal/storage"
	"m4jordomo/internal/types"
	"sync"
)

type DevicesPlugin struct {
	mu      sync.RWMutex
	states  map[string]bool
	storage *storage.Storage
}

var defaultDevices = map[string]bool{
	"light":   false,
	"heating": false,
	"door":    false,
}

func New(st *storage.Storage) *DevicesPlugin {
	return &DevicesPlugin{
		states:  make(map[string]bool),
		storage: st,
	}
}

func (p *DevicesPlugin) Name() string {
	return "devices"
}

func (p *DevicesPlugin) Init(b *bus.Bus) error {
	// 1. Пытаемся загрузить состояния из БД
	if err := p.loadStatesFromDB(); err != nil {
		log.Printf("[Devices] ⚠️ Не удалось загрузить состояния: %v", err)
	}

	// 2. Добавляем недостающие устройства (если БД пустая или старая)
	p.mu.Lock()
	for name, status := range defaultDevices {
		if _, exists := p.states[name]; !exists {
			p.states[name] = status
			p.storage.SetDeviceStatus(name, status)
		}
	}
	p.mu.Unlock()

	// 3. Проверяем, загрузились ли состояния. Если нет (все еще пусто) — создаем дефолтные принудительно
	p.mu.RLock()
	empty := len(p.states) == 0
	p.mu.RUnlock()

	if empty {
		log.Println("[Devices] База данных пуста. Создаю дефолтные устройства...")
		p.setDefaults()
	}

	log.Printf("[Devices] 📋 Загружены состояния: %v", p.states)

	log.Println("[Devices] Подписываюсь на события...")

	b.Subscribe("command.device.set", func(e types.Event) {
		p.handleSetDevice(e, b)
	})

	b.Subscribe("command.device.get_all", func(e types.Event) {
		p.handleGetAllDevices(e, b)
	})

	log.Println("[Devices] Подписки выполнены.")

	return nil
}

func (p *DevicesPlugin) loadStatesFromDB() error {
	states, err := p.storage.GetAllDevices()
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range states {
		p.states[s.Name] = s.Status
	}
	return nil
}

func (p *DevicesPlugin) setDefaults() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, status := range defaultDevices {
		p.states[name] = status
		// Сохраняем в БД
		if err := p.storage.SetDeviceStatus(name, status); err != nil {
			log.Printf("[Devices] Ошибка сохранения %s: %v", name, err)
		}
	}
	log.Println("[Devices] ✅ Дефолтные устройства созданы и сохранены в БД")
}

func (p *DevicesPlugin) handleSetDevice(e types.Event, bus *bus.Bus) {
	deviceName, ok := e.Payload["name"].(string)
	if !ok {
		log.Printf("[Devices] Ошибка: нет поля 'name'")
		return
	}
	status, ok := e.Payload["status"].(bool)
	if !ok {
		log.Printf("[Devices] Ошибка: нет поля 'status'")
		return
	}
	p.mu.Lock()
	_, exists := p.states[deviceName]
	if !exists {
		p.mu.Unlock()
		log.Printf("[Devices] ❌ Устройство '%s' не найдено", deviceName)
		return
	}
	p.states[deviceName] = status
	p.mu.Unlock()

	if err := p.storage.SetDeviceStatus(deviceName, status); err != nil {
		log.Printf("[Devices] ❌ Ошибка сохранения в БД: %v", err)
	}
	log.Printf("[Devices] 💡 %s теперь %v", deviceName, status)

	bus.Publish(types.Event{
		Type:     "device.state.changed",
		Priority: types.Medium,
		Payload: map[string]interface{}{
			"name":   deviceName,
			"status": status,
		},
	})
}

func (p *DevicesPlugin) handleGetAllDevices(e types.Event, bus *bus.Bus) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	bus.Publish(types.Event{
		Type:     "device.state.response",
		Priority: types.Medium,
		Payload: map[string]interface{}{
			"states": p.states,
		},
	})
}
