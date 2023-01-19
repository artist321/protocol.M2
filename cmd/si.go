/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// siCmd represents the si command
var siCmd = &cobra.Command{
	Use:   "si",
	Short: "Интерфейс для работы с разделом Утверждённые типы средств измерений (ФИФ)",
	Long: `Интерфейс для работы с разделом Утверждённые типы средств измерений (ФИФ)

-f -- обновление всей базы из ФИФ
-new -- обновление только последний 200 номеров ГРСИ
-init -- инициализация локальной базы ГРСИ

`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("si called")
	},
}

func init() {
	rootCmd.AddCommand(siCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// siCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// siCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
