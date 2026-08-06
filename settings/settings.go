package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type SettingsType struct {
	ServerName  string `json:"servername"`
	ServerPort  string `json:"serverport"`
	Topik       string `json:"topik"`
	URL         string `json:"url"`
	TmpFolder   string `json:"tmpfolder"`
	LogFilePath string `json:"logfilepath"`
}

var (
	logFileName      = "ntfy-speaker"
	logFileExtension = ".log"
)

func (s *SettingsType) New() {
	// Парсим аргументы командной строки
	pflag.StringVarP(&s.ServerName, "server", "s", "http://localhost", "Сервер подключения (http://localhost)")
	pflag.StringVarP(&s.ServerPort, "port", "p", "80", "Порт подключения")
	pflag.StringVarP(&s.Topik, "topik", "t", "test-topik", "Топик")

	pflag.Parse()

	// Формируем URL для подключения
	if s.ServerPort != "" {
		s.URL = fmt.Sprintf("%s:%s/%s/json", s.ServerName, s.ServerPort, s.Topik)
	} else {
		s.URL = fmt.Sprintf("%s/%s/json", s.ServerName, s.Topik)
	}

	// Создаем временную папку для иконок и логов
	s.TmpFolder = filepath.Join(os.TempDir(), "ntfy")
	os.MkdirAll(s.TmpFolder, 0755) // Используем 0755 вместо 0644 для директорий

	// Формируем путь к лог-файлу
	srtLogFileName := logFileName + "_" + s.Topik + logFileExtension
	s.LogFilePath = filepath.Join(s.TmpFolder, srtLogFileName)
}
