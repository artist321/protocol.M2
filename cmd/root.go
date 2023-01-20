/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"
)

var (
	Version string
	Build   string
)

var (
	v string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "protocol.M2",
	Short: "Многофункциональная программа для метрологов",
	Long: `Многофункциональная программа для метрологов Protocol.M2

Ничего НЕ ГАРАНТИРУЕТСЯ, в пределах, ограниченных законом.

Сообщения об ошибках и вопросы отправляйте на <a.a.demchenko@yandex.com>
Чтобы узнать как пользоваться программой, введите:	protocol.M2 help
Версия: ` + Version + Build,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

	},
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if err := setupLog(v); err != nil {
			return err
		}
		return nil
	},
}

func setupLog(level string) error {
	logger := &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out: io.MultiWriter(os.Stdout),
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	logrus.SetOutput(logger.Writer())
	logrus.SetFormatter(logger.Formatter)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.PersistentFlags().StringVarP(
		&v, "verbosity", "v", logrus.WarnLevel.String(), "Log level (debug, info, warn, error, fatal, panic",
	)
	rootCmd.Flags().BoolP("help", "h", false, "Help message for help?")
}
