package speaker

import (
	_ "embed"
	"fmt"
	pl "ntfy-speaker/listener"
	"ntfy-speaker/settings"
	"os"
	"path/filepath"

	"ntfy-speaker/toast"
)

// https://icon-icons.com/authors/1104-afshin-t2y
// https://icon-icons.com/pack/iconly-v23-bulk/2945
//
//go:embed src/wtf-icon.png
var embeddedIcon_wtf []byte

//go:embed src/gooden-icon.png
var embeddedIcon_gooden []byte

//go:embed src/information-icon.png
var embeddedIcon_info []byte

//go:embed src/Hay-icon.png
var embeddedIcon_warn []byte

//go:embed src/warning-icon.png
var embeddedIcon_err []byte

// Глобальная переменная для хранения пути к извлеченной иконке
var tempIconPath_wtf string
var tempIconPath_gooden string
var tempIconPath_info string
var tempIconPath_warn string
var tempIconPath_err string

// initIcon извлекает картинку во временную папку только один раз
func initIcon(settings *settings.SettingsType) error {
	// Если путь уже определен, значит мы уже извлекали файл
	if tempIconPath_wtf != "" && tempIconPath_gooden != "" && tempIconPath_info != "" && tempIconPath_warn != "" && tempIconPath_err != "" {
		return nil
	} else {
		tempIconPath_wtf = filepath.Join(settings.TmpFolder, "wtf-icon.png")
		tempIconPath_gooden = filepath.Join(settings.TmpFolder, "gooden-icon.png")
		tempIconPath_info = filepath.Join(settings.TmpFolder, "information-icon.png")
		tempIconPath_warn = filepath.Join(settings.TmpFolder, "Hay-icon.png")
		tempIconPath_err = filepath.Join(settings.TmpFolder, "warning-icon.png")

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_wtf); err == nil {

		} else {
			// Записываем встроенные байты в файл
			err := os.WriteFile(tempIconPath_wtf, embeddedIcon_wtf, 0644)
			if err != nil {
				return err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_gooden); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_gooden, embeddedIcon_gooden, 0644)
			if err != nil {
				return err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_info); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_info, embeddedIcon_info, 0644)
			if err != nil {
				return err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_warn); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_warn, embeddedIcon_warn, 0644)
			if err != nil {
				return err
			}
		}

		// Проверяем, существует ли уже файл (чтобы не перезаписывать его лишний раз)
		if _, err := os.Stat(tempIconPath_err); err == nil {

		} else {
			err = os.WriteFile(tempIconPath_err, embeddedIcon_err, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Speak(msg pl.NtfyMessageType, settings *settings.SettingsType) error {

	err := initIcon(settings)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации иконок push-уведомлений: %w", err)
	}

	// Кнопки
	var action []toast.Action

	notification := toast.Notification{
		AppID:   "NtfySpeaker - " + msg.Topic, // Идентификатор вашего приложения
		Title:   msg.Title,
		Message: msg.Message,
	}

	// При нажатии
	if msg.Click != "" {
		notification.ActivationArguments = settings.ServerName + ":" + settings.ServerPort + "/" + settings.Topik
		action = append(action, toast.Action{
			Type:      "protocol",
			Label:     "Ссылка URL",
			Arguments: msg.Click,
		})
	} else {
		notification.ActivationArguments = settings.ServerName + ":" + settings.ServerPort + "/" + settings.Topik
	}

	// Icon и Duration ("short" или "long")
	switch msg.Priority {
	case 0: // Средний приоритет
		notification.Icon = tempIconPath_info
		notification.Duration = "long"

	case 1: // Низкий приоритет
		notification.Duration = "short"
		notification.Icon = tempIconPath_wtf
	case 2: // Низкий приоритет
		notification.Duration = "short"
		notification.Icon = tempIconPath_gooden
	case 3: // Средний приоритет
		notification.Icon = tempIconPath_info
		notification.Duration = "long"
	case 4: // Высокий приоритет
		notification.Icon = tempIconPath_warn
		notification.Scenario = "reminder"
	case 5: // Высокий приоритет
		notification.Icon = tempIconPath_err
		notification.Scenario = "reminder"
	}

	if msg.Attachment != nil {
		action = append(action, toast.Action{
			Type:      "protocol",
			Label:     msg.Attachment.Name,
			Arguments: msg.Attachment.URL,
		})
	}

	notification.Actions = action

	// Отправка
	err = notification.Push()
	if err != nil {
		return fmt.Errorf("ошибка при отправке уведомления в панель уведомлений Windows %w", err)
	}

	return nil
}
