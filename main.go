/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package main

import (
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"protocol.M2/cmd"
	"protocol.M2/utils"
	"runtime"
	"time"
)

const dirApp = "protocol.M2"

func main() {

	cmd.Execute()
}

func init() {
	rand.Seed(time.Now().Unix() / 1986)

	//logrus.SetLevel(logrus.InfoLevel)
	//logrus.SetLevel(logrus.DebugLevel)

	switch runtime.GOOS {
	case "windows":
		{
			PathSI := path.Join(utils.UserHomeDir(), dirApp)
			if err := utils.EnsureMakeDir(PathSI); err != nil {
				logrus.Errorln("folder creation failed with:" + err.Error())
				os.Exit(1)
			}
		}
	case "darwin":
		{
			PathSI := path.Join(utils.UserHomeDir(), dirApp)
			if err := utils.EnsureMakeDir(PathSI); err != nil {
				logrus.Errorln("folder creation failed with:" + err.Error())
				os.Exit(1)
			}
		}
	case "linux":
		{
			PathSI := path.Join(utils.UserHomeDir(), dirApp)
			if err := utils.EnsureMakeDir(PathSI); err != nil {
				logrus.Errorln("folder creation failed with:" + err.Error())
				os.Exit(1)
			}
		}
	} // создаем каталог для работы
}
