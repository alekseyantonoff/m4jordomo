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
				Type:     "command.device.get_all",
				Priority: types.Medium,
				Payload:  map[string]interface{}{},
			})
		case strings.Contains(text, "включи свет"):
			bus.Publish(types.Event{
				Type:     "command.device.set",
				Priority: types.High,
				Payload: map[string]interface{}{
					"name":   "light",
					"status": true,
				},
			})
		case strings.Contains(text, "выключи свет"):
			bus.Publish(types.Event{
				Type:     "command.device.set",
				Priority: types.High,
				Payload: map[string]interface{}{
					"name":   "light",
					"status": false,
				},
			})
		case strings.Contains(text, "включи отопление"):
			bus.Publish(types.Event{
				Type:     "command.device.set",
				Priority: types.High,
				Payload: map[string]interface{}{
					"name":   "heating",
					"status": true,
				},
			})
		case strings.Contains(text, "выключи отопление"):
			bus.Publish(types.Event{
				Type:     "command.device.set",
				Priority: types.High,
				Payload: map[string]interface{}{
					"name":   "heating",
					"status": false,
				},
			})
		default:
			log.Printf("[Voice] 🤔 Не понял команду: '%s'", text)
		}
	}
}
