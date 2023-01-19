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
	//"log"
	"os"
	"path"
	"path/filepath"
	"protocol.M2/log"
	"protocol.M2/utils"
	"strings"
)

// oeiCmd represents the oei command
var oeiCmd = &cobra.Command{
	Use:   "oei",
	Short: "Интерфейс для работы с oei-analitika.ru",
	Long: `Интерфейс для работы с oei-analitika.ru

-new -- обновление файлов из раздела Полезные документы
`,
	Run: func(cmd *cobra.Command, args []string) {
		ScrapFilesFromOEI()
	},
}

func init() {
	rootCmd.AddCommand(oeiCmd)
	//Log.Println("")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	oeiCmd.PersistentFlags().BoolP("new", "n", false, "Обновление файлов из раздела Полезные документы")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// oeiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ScrapFilesFromOEI() {
	var noticeFmt = color.New(color.FgGreen).PrintlnFunc()
	noticeFmt("[oei-analitika] Старт.")
	var urls, p0 []string
	fName := "OEI-A-RE_data.csv"
	fn := ""
	link := ""

	f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Log.Error("Файл cvs не создан, ошибка: %q", err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"),
		//colly.Async(false),
	)
	c.OnRequest(
		func(r *colly.Request) {
			r.Headers.Set("Accept", "*/*")
			r.Headers.Set("Cookie", "PHPSESSID=7q943qaafa4hng9s0geci2te9k")
		},
	)
	c.OnResponse(
		func(r *colly.Response) {
			//log.Println("Got: ", r.Request.URL)
		},
	)
	c.OnHTML(
		"table[id=good_docs]", func(e *colly.HTMLElement) {
			//fmt.Println(e)
			e.ForEach(
				"tr", func(_ int, el *colly.HTMLElement) {
					//fmt.Println(el.ChildText("td:nth-child(8)"))
					el.ForEach(
						"td", func(_ int, em *colly.HTMLElement) {

							em.ForEach(
								"a", func(_ int, en *colly.HTMLElement) {
									link = en.Attr("href")
									//urls = append(urls, link)
									//fmt.Println("h", en.Attr("href"))
								},
							)
						},
					)
					w.Write(
						[]string{
							//el.ChildText("td:nth-child(1)"),
							el.ChildText("td:nth-child(2)"),
							el.ChildText("td:nth-child(3)"),
							el.ChildText("td:nth-child(4)"),
							el.ChildText("td:nth-child(5)"),
							el.ChildText("td:nth-child(6)"),
							el.ChildText("td:nth-child(7)"),
							link,
							el.ChildText("td:nth-child(9)"),
						},
					)
					switch el.ChildText("td:nth-child(5)") {
					case "Методика поверки":
						{
							fn = fmt.Sprintf("%s", "mp_")
							if len(el.ChildText("td:nth-child(3)")) == 0 {
								fn = fn + el.ChildText("td:nth-child(2)")
							} else {
								fn = fn + strings.ReplaceAll(el.ChildText("td:nth-child(3)"), "/", "_")
							}
						}
					case "Паспорт":
						{
							fn = fmt.Sprintf("%s", "passp_")
							if len(el.ChildText("td:nth-child(3)")) == 0 {
								fn = fn + el.ChildText("td:nth-child(2)")
							} else {
								fn = fn + strings.ReplaceAll(el.ChildText("td:nth-child(3)"), "/", "_")
							}
						}

					case "Руководство по эксплуатации":
						{
							fn = fmt.Sprintf("%s", "re_")
							if len(el.ChildText("td:nth-child(3)")) == 0 {
								fn = fn + el.ChildText("td:nth-child(2)")
							} else {
								fn = fn + strings.ReplaceAll(el.ChildText("td:nth-child(3)"), "/", "_")
							}
						}
					case "Описание типа":
						{
							fn = fmt.Sprintf("%s", "ot_")
							if len(el.ChildText("td:nth-child(3)")) == 0 {
								fn = fn + el.ChildText("td:nth-child(2)")
							} else {
								fn = fn + strings.ReplaceAll(el.ChildText("td:nth-child(3)"), "/", "_")
							}
						}
					}
					//fmt.Println(el.ChildText("td:nth-child(3)"))

					link = fmt.Sprintf("http://oei-analitika.ru/kurilka/%s", link)
					ext := filepath.Ext(link)
					urls = append(urls, link)

					p := path.Join("OEI", strings.ReplaceAll(fmt.Sprintf("%s%s", fn, ext), "/", "_"))
					p0 = append(p0, p)
					//DownloadFileFromFif(link, p)
					//time.Sleep(333 * time.Millisecond)
				},
			)
			noticeFmt("сканирование завершено")
		},
	)
	noticeFmt("загружаю данные...")
	c.OnError(
		func(r *colly.Response, err error) {
			log.Log.Info("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		},
	)
	c.Visit(fmt.Sprintf("http://oei-analitika.ru/kurilka/reestr_good_docs.php"))
	w.Flush()
	for i := len(urls) - 1; i >= 0; i-- {
		err := utils.DownloadFile(urls[i], p0[i])
		if err != nil {
			log.Log.Error(err)
			continue
		}
		//fmt.Println(urls[i])
	}

	//noticeFmt("[oei-analitika] скачиваем.")
	//MultiDownloader(urls)

	noticeFmt("[oei-analitika] Выполнено.")
}
