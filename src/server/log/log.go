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
	return []byte(fmt.Sprintf("%s %s %s\n", timestamp, f.LevelDesc[entry.Level], entry.Message)), nil
}

var log *logrus.Logger

func init() {
	log = logrus.New()
	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006/01/02 15:04:05"
	plainFormatter.LevelDesc = []string{"[panic  ]", "[fetal  ]", "[error  ]", "[warn   ]", "[info   ]", "[debug  ]"}
	log.SetFormatter(plainFormatter)
	log.SetLevel(logrus.DebugLevel)
}

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
		fmt.Println("[SendToLogServer] log msg marshal error")
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(logMsgStr))
	if err != nil {
		fmt.Println("[SendToLogServer] req 错误")
	} else {
		req.Header.Set("Content-Type", "application/json")
		client := http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("[SendToLogServer] client 请求错误")
		}

		if resp == nil {
			fmt.Println("[SendToLogServer] resp为空")
			return
		} else if resp.StatusCode != 200 {
			bs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("[SendToLogServer]")
			}

			fmt.Println("[SendToLogServer]" + string(bs))
			return
		}
	}
}
