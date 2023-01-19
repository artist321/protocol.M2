/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// fsaCmd represents the fsa command
var fsaCmd = &cobra.Command{
	Use:   "fsa",
	Short: "Интерфейс для работы с fsa.gov.ru",
	Long:  `Интерфейс для работы с fsa.gov.ru`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fsa called")
	},
}

func init() {
	rootCmd.AddCommand(fsaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fsaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fsaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
