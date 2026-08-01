// Плагин голосового управления

package voice

import (
	"bufio"
	"log"
	"m4jordomo/internal/bus"
	"m4jordomo/internal/types"
	"os"
	"strings"
)

type VoicePlugin struct{}

func New() *VoicePlugin {
	return &VoicePlugin{}
}

func (p *VoicePlugin) Name() string {
	return "voice"
}

func (p *VoicePlugin) Init(bus *bus.Bus) error {
	log.Println("[Voice] 🎤 Запущен голосовой интерфейс (вводите команды в консоль)")
	log.Println("[Voice] Доступные команды:")
	log.Println("  - включи свет")
	log.Println("  - выключи свет")
	log.Println("  - включи отопление")
	log.Println("  - выключи отопление")
	log.Println("  - открой дверь")
	log.Println("  - закрой дверь")
	log.Println("  - статус")
	log.Println("  - выход")
	go p.listenConsole(bus)
	return nil
}

func (p *VoicePlugin) listenConsole(bus *bus.Bus) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("[Voice] Ошибка чтения: %v", err)
			continue
		}
		text = strings.TrimSpace(strings.ToLower(text))

		switch {
		case text == "выход":
			log.Println("[Voice] 👋 Завершение работы...")
			os.Exit(0)
		case text == "статус":
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceGetAll,
				Priority: types.Medium,
				Payload:  nil,
			})
		case strings.Contains(text, "включи свет"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "light",
					Status: true,
				},
			})
		case strings.Contains(text, "выключи свет"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "light",
					Status: false,
				},
			})
		case strings.Contains(text, "включи отопление"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "heating",
					Status: true,
				},
			})
		case strings.Contains(text, "выключи отопление"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "heating",
					Status: false,
				},
			})

		case strings.Contains(text, "открой дверь"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "door",
					Status: true,
				},
			})
		case strings.Contains(text, "закрой дверь"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "door",
					Status: false,
				},
			})

		default:
			log.Printf("[Voice] 🤔 Не понял команду: '%s'", text)
		}
	}
}
