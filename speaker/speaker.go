package speaker

import (
	_ "embed"
	"fmt"
	pl "ntfy_speaker/listener"
	"ntfy_speaker/settings"
	"os"
	"path/filepath"

	"github.com/go-toast/toast"
)

//go:embed src/information-icon.png
var embeddedIcon_info []byte

//go:embed src/plus-circle-icon.png
var embeddedIcon_plus []byte

//go:embed src/warning-icon.png
var embeddedIcon_warn []byte

// Глобальная переменная для хранения пути к извлеченной иконке
var tempIconPath_info string
var tempIconPath_plus string
var tempIconPath_warn string

// initIcon извлекает картинку во временную папку только один раз
func initIcon() (string, error) {
	// Если путь уже определен, значит мы уже извлекали файл
	if tempIconPath_info != "" && tempIconPath_plus != "" && tempIconPath_warn != "" {
		return "", nil
	} else {
		// Получаем системную временную директорию (обычно C:\Users\Username\AppData\Local\Temp)
		tempDir := os.TempDir() + "\\ntfy"
		os.Mkdir(tempDir, 0644)

		fmt.Println(tempDir)

		tempIconPath_info = filepath.Join(tempDir, "information-icon.png")
		tempIconPath_plus = filepath.Join(tempDir, "plus-circle-icon.png")
		tempIconPath_warn = filepath.Join(tempDir, "warning-icon.png")

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_info); err == nil {

		} else {
			// Записываем встроенные байты в файл
			err := os.WriteFile(tempIconPath_info, embeddedIcon_info, 0644)
			if err != nil {
				return tempDir, err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_plus); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_plus, embeddedIcon_plus, 0644)
			if err != nil {
				return tempDir, err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_warn); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_warn, embeddedIcon_warn, 0644)
			if err != nil {
				return tempDir, err
			}
		}
	}

	return "", nil
}

func Speak(msg pl.NtfyMessageType, settings *settings.SettingsType) error {

	initIcon()

	notification := toast.Notification{
		AppID:   "NtfySpeaker - " + msg.Topic, // Идентификатор вашего приложения
		Title:   msg.Title,
		Message: msg.Message,
	}

	// При нажатии
	if msg.Click != "" {
		notification.ActivationArguments = msg.Click
	} else {
		notification.ActivationArguments = settings.ServerName + ":" + settings.ServerPort
	}

	// Icon и Duration ("short" или "long")
	if msg.Priority < 3 && msg.Priority != 0 { // Низкий приоритет
		notification.Duration = "short"
		notification.Icon = tempIconPath_plus
	} else if msg.Priority > 3 && msg.Priority != 0 { // Средний приоритет
		notification.Duration = "long"
		notification.Icon = tempIconPath_warn
	} else {
		notification.Icon = tempIconPath_info
	}

	// Кнопки
	var action []toast.Action

	if msg.Attachment != nil {
		action = append(action, toast.Action{
			Type:      "protocol",
			Label:     msg.Attachment.Name,
			Arguments: msg.Attachment.URL,
		})
	}

	notification.Actions = action

	// Отправка
	err := notification.Push()
	if err != nil {
		return err
	}

	return nil
}
