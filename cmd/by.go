/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"os"
	"path"
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
		ScrapBelGRSI()
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

func ScrapBelGRSI() {
	var noticeFmt = color.New(color.FgGreen).PrintlnFunc()
	noticeFmt("[Belarus GRSI] Старт.")
	fName := "BelGRSI_SI.csv"
	file, err := os.OpenFile(fName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Файл csv не создан, ошибка: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// make dir for Bel.GRSI files
	utils.EnsureMakeDir(path.Join(utils.RootDir, "BelGRSI"))

	for i := 1; i < 441; i++ {
		c := colly.NewCollector()
		c.WithTransport(
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		)
		c.OnHTML(
			"div[id=w10-container]", func(e *colly.HTMLElement) {
				e.ForEach(
					"tr[class=w10]", func(_ int, el *colly.HTMLElement) {
						//fmt.Println(el.Attr("data-key"))
						y, _ := strconv.Atoi(el.Attr("data-key"))
						ScrapBelGRSIInner(y)
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
				fmt.Println(i, "Сканирование завершено")
			},
		)
		c.Visit(fmt.Sprintf("https://www.oei.by/grsi/index?page=%d&per-page=100&sort=-grsi_date", i))
		writer.Flush()
		//time.Sleep(0 * time.Millisecond)
	}
	noticeFmt("[Belarus GRSI] Выполнено.")

}

func ScrapBelGRSIInner(i int) {
	//fmt.Printf("https://www.oei.by/grsi/view?id=%d\n", i)
	fName := "BelGRSI_data_si.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Файл не создан, ошибка: %q", err)
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
								err := utils.DownloadFile(em.Attr("href"), path.Join("BelGRSI", em.Text))
								if err != nil {
									log.Error(err)
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
						log.Fatalf("ошибка writer.Write: %q", err)
						return
					}
				},
			)
			//fmt.Println(i, "Scrapping Complete")
		},
	)
	err = c.Visit(fmt.Sprintf("https://www.oei.by/grsi/view?id=%d", i))
	if err != nil {
		log.Error(err)
	}
	time.Sleep(0 * time.Millisecond)
}
