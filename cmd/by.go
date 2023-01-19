/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// byCmd represents the by command
var byCmd = &cobra.Command{
	Use:   "by",
	Short: "Интерфейс для работы с Государственным реестром средств измерений Респ.Беларусь",
	Long: `Интерфейс для работы с Государственным реестром средств измерений Респ.Беларусь
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("by called")
	},
}

func init() {
	rootCmd.AddCommand(byCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// byCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// byCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
