package conf

import (
	"encoding/json"
	"io/ioutil"

	"github.com/name5566/leaf/log"
)

var Server struct {
	LogLevel    string
	LogPath     string
	WSAddr      string
	CertFile    string
	KeyFile     string
	TCPAddr     string
	MaxConnNum  int
	ConsolePort int
	ProfilePath string

	TokenServer      string
	CenterServer     string
	CenterServerPort string
	DevKey           string
	DevName          string
	GameID           string
	MongoDB          string
}

func init() {
	// 配置文件
	fileName := "conf/server.json"
	log.Debug("读取配置文件 %v...", fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v", err)
	}
}
