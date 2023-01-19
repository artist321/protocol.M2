/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// oeiCmd represents the oei command
var oeiCmd = &cobra.Command{
	Use:   "oei",
	Short: "Интерфейс для работы с oei-analitika.ru",
	Long: `Интерфейс для работы с oei-analitika.ru

-new -- обновление файлов из раздела Полезные документы
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("oei called")
	},
}

func init() {
	rootCmd.AddCommand(oeiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// oeiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// oeiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
