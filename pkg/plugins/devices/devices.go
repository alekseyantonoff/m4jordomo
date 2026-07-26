// Плагин управления устройствами
package devices

import (
	"log"
	"sync"
	"m4jordomo/internal/storage"
	"m4jordomo/internal/types"
)

type DevicesPlugin struct {
	mu      sync.RWMutex
	states  map[string]bool
	storage *storage.Storage
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

func (p *DevicesPlugin) Init(bus *types.Bus) error {
	if err := p.loadStatesFromDB(); err != nil {
		log.Printf("[Devices] ⚠️ Не удалось загрузить состояния: %v", err)
		p.setDefaults()
	}
	log.Printf("[Devices] 📋 Загружены состояния: %v", p.states)

	bus.Subscribe("command.device.set", func(e types.Event) {
		p.handleSetDevice(e, bus)
	})
	bus.Subscribe("command.device.get_all", func(e types.Event) {
		p.handleGetAllDevices(e, bus)
	})
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
	defaults := map[string]bool{"light": false, "heating": false, "door": false}
	for name, status := range defaults {
		p.states[name] = status
		if err := p.storage.SetDeviceStatus(name, status); err != nil {
			log.Printf("[Devices] Ошибка сохранения %s: %v", name, err)
		}
	}
}

func (p *DevicesPlugin) handleSetDevice(e types.Event, bus *types.Bus) {
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

func (p *DevicesPlugin) handleGetAllDevices(e types.Event, bus *types.Bus) {
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

