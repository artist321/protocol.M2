/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package log

import (
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"
)

var flog = "protocol.M2.log"
var Log *logrus.Logger

func init() {
	f, err := os.OpenFile(flog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logrus.Error("Ошибка создания logfile" + flog)
		panic(err)
	}
	defer f.Close()

	Log = &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel, //InfoLevel

		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}
	Log.Println()
}
