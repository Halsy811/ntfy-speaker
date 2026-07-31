package settings

import (
	"fmt"

	"github.com/spf13/pflag"
)

type SettingsType struct {
	ServerName string `json:"servername"`
	ServerPort string `json:"serverport"`
	Topik      string `json:"topik"`
	URL        string `json:"url"`
}

func (s *SettingsType) New() {
	pflag.StringVarP(&s.ServerName, "server", "s", "http://localhost", "Сервер подключения (http://localhost)")
	pflag.StringVarP(&s.ServerPort, "port", "p", "80", "Порт подключения")
	pflag.StringVarP(&s.Topik, "topik", "t", "test_zm", "Топик")

	pflag.Parse()
	if s.ServerPort != "" {
		s.URL = fmt.Sprintf("%s:%s/%s/json", s.ServerName, s.ServerPort, s.Topik)
	} else {
		s.URL = fmt.Sprintf("%s/%s/json", s.ServerName, s.Topik)
	}
}
