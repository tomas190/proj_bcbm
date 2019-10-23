package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"proj_bcbm/src/server/conf"
	"time"
)

type PlainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

// LogCenter 日志中心数据结构
type MsgLogServer struct {
	Type     string `json:"type"`      //"LOG"|"ERR"|"DEG",
	From     string `json:"from"`      //"game-server",
	GameName string `json:"game_name"` // "lunpan"
	Host     string `json:"host"`      //服务IP地址,
	Msg      string `json:"msg"`
	Time     string `json:"time"` // 时间(YYYY-MM-DD HH:II:SS),
}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("%s %s %s\n", f.LevelDesc[entry.Level], timestamp, entry.Message)), nil
}

func init() {
	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	plainFormatter.LevelDesc = []string{"PANC", "FATL", "ERRO", "WARN", "INFO", "DEBG"}
	log.SetFormatter(plainFormatter)
	log.SetLevel(logrus.DebugLevel)
}

var log = logrus.New()

func Debug(format string, a ...interface{}) {
	log.Debugf(format, a...)
	go SendToLogServer("DEG", fmt.Sprintf(format, a...), time.Now().Format("2006-01-02 15:04:05"))
}

func Error(format string, a ...interface{}) {
	log.Errorf(format, a...)
	go SendToLogServer("ERR", fmt.Sprintf(format, a...), time.Now().Format("2006-01-02 15:04:05"))
}

func Fatal(format string, a ...interface{}) {
	log.Fatalf(format, a...)
	go SendToLogServer("ERR", fmt.Sprintf(format, a...), time.Now().Format("2006-01-02 15:04:05"))

}

func SendToLogServer(t string, msg string, timeStr string) {
	url := conf.Server.LogServer

	logMsg := MsgLogServer{
		Type:     t,
		From:     "game-server",
		GameName: "benchibaoma",
		Host:     "",
		Msg:      msg,
		Time:     timeStr,
	}

	logMsgStr, err := json.Marshal(&logMsg)
	if err != nil {
		fmt.Println("log msg marshal error")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(logMsgStr))
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)

	if resp == nil || resp.StatusCode != 200 {
	} else {
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("响应体读取失败", err)
		}

		fmt.Println(string(bs))
	}
}
