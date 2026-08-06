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

const logFileName = "ntfy-speaker.log"

func (s *SettingsType) New() {
	pflag.StringVarP(&s.ServerName, "server", "s", "http://localhost", "Сервер подключения (http://localhost)")
	pflag.StringVarP(&s.ServerPort, "port", "p", "80", "Порт подключения")
	pflag.StringVarP(&s.Topik, "topik", "t", "test-topik", "Топик")

	pflag.Parse()
	if s.ServerPort != "" {
		s.URL = fmt.Sprintf("%s:%s/%s/json", s.ServerName, s.ServerPort, s.Topik)
	} else {
		s.URL = fmt.Sprintf("%s/%s/json", s.ServerName, s.Topik)
	}

	// Получаем системную временную директорию (обычно C:\Users\UserName\AppData\Local\Temp)
	s.TmpFolder = filepath.Join(os.TempDir(), "ntfy")
	os.Mkdir(s.TmpFolder, 0644)

	s.LogFilePath = filepath.Join(s.TmpFolder, logFileName)
}
