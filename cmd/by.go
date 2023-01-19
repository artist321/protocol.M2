/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cobra"
	"os"
	"path"
	"protocol.M2/log"
	"protocol.M2/utils"
	"strconv"
	"time"
)

// byCmd represents the by command
var byCmd = &cobra.Command{
	Use:   "by",
	Short: "Интерфейс для работы с Государственным реестром средств измерений Респ.Беларусь",
	Long: `Интерфейс для работы с Государственным реестром средств измерений Респ.Беларусь

--update, -u -- Обновление ГРСИ (Респ.Белорусь)
`,
	Run: func(cmd *cobra.Command, args []string) {
		ScrapByGRSI()
	},
}

func init() {
	rootCmd.AddCommand(byCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	byCmd.PersistentFlags().BoolP("update", "u", false, "Обновление ГРСИ (Респ.Белорусь)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// byCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ScrapByGRSI() {
	var noticeFmt = color.New(color.FgGreen).PrintlnFunc()
	noticeFmt("[by-grsi] Старт.")
	fName := "ByGRSI_data.csv"
	file, err := os.OpenFile(fName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Log.Fatalf("Файл csv не создан, ошибка: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 1; i < 441; i++ {
		c := colly.NewCollector()
		c.OnHTML(
			"div[id=w10-container]", func(e *colly.HTMLElement) {
				e.ForEach(
					"tr[class=w10]", func(_ int, el *colly.HTMLElement) {
						//fmt.Println(el.Attr("data-key"))

						y, _ := strconv.Atoi(el.Attr("data-key"))
						ScrapByGRSIInner(y)
						writer.Write(
							[]string{
								//el.ChildText("td:nth-child(1)"),
								el.ChildText("td:nth-child(2)"),
								el.ChildText("td:nth-child(3)"),
								el.ChildText("td:nth-child(4)"),
								el.ChildText("td:nth-child(5)"),
								el.ChildText("td:nth-child(6)"),
								el.ChildText("td:nth-child(7)"),
								el.ChildText("td:nth-child(8)"),
								el.ChildText("td:nth-child(9)"),
								el.ChildText("td:nth-child(10)"),
								el.ChildText("td:nth-child(11)"),
							},
						)
					},
				)
				fmt.Println(i, "Scrapping Complete")
			},
		)
		c.Visit(fmt.Sprintf("https://www.oei.by/grsi/index?page=%d&per-page=100&sort=-grsi_date", i))
		writer.Flush()
		//time.Sleep(0 * time.Millisecond)
	}
	noticeFmt("[by-grsi] Выполнено.")

}

func ScrapByGRSIInner(i int) {
	//fmt.Printf("https://www.oei.by/grsi/view?id=%d\n", i)
	fName := "ByGRSI_data_si.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Log.Fatalf("Файл не создан, ошибка: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	c := colly.NewCollector()
	c.OnHTML(
		"div.col-lg-6.txt-fields", func(e *colly.HTMLElement) {
			//fmt.Println(e.Text)
			//fmt.Println(c.Wait)
			e.ForEach(
				"div.form-group", func(_ int, el *colly.HTMLElement) {
					//fmt.Println(el.Text)

					if len(el.ChildText("div.form-group > p > a")) > 0 {
						el.ForEach(
							"div.form-group > p > a", func(_ int, em *colly.HTMLElement) {
								//fmt.Println("h", em.Attr("href"))
								//fmt.Println("t", em.Text)
								//fmt.Println(em.Text)
								err := utils.DownloadFile(em.Attr("href"), path.Join("ByGRSI", em.Text))
								if err != nil {
									log.Log.Error(err)
								}
							},
						)
					}
					err := writer.Write(
						[]string{
							el.ChildText("div.form-group > p"),
						},
					)
					if err != nil {
						log.Log.Fatalf("ошибка writer.Write: %q", err)
						return
					}
				},
			)
			//fmt.Println(i, "Scrapping Complete")
		},
	)
	//c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	//	link := e.Attr("href")
	//	if !strings.HasPrefix(link, "http://media.belgim.by/grsi/") {
	//		return
	//	}
	//	//DownloadFileFromFif(link)
	//	// start scraping the page under the link found
	//	e.Request.Visit(link)
	//})
	err = c.Visit(fmt.Sprintf("https://www.oei.by/grsi/view?id=%d", i))
	if err != nil {
		log.Log.Error(err)
	}
	time.Sleep(0 * time.Millisecond)
}
