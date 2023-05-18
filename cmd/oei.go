/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package cmd

import (
	// "container/list"

	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"protocol.M2/utils"
)

// OEIDesc contains a url, description for a project.
type OEIDesc struct {
	URL, Description string
}

// TRows contains rows of table.
type TRows struct {
	Num      string
	DocName  string
	DocDesc  string
	Author   string
	DocType  string
	WhoUp    string
	DateUp   string
	Link     string
	TypeSize string
}

// oeiCmd represents the "oei" command
var oeiCmd = &cobra.Command{
	Use:   "oei",
	Short: "Интерфейс для работы с oei-analitika.ru",
	Long: `Интерфейс для работы с oei-analitika.ru

--new, -n -- обновление файлов из раздела Полезные документы
`,
	Run: func(cmd *cobra.Command, args []string) {
		ScrapFilesFromOEIv5()
	},
}

var (
	CHROME_EXEC_PATH  = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	USER_DATA_DIR     = fmt.Sprintf("/Users/%s/Library/Application Support/Google/Chrome/", utils.GetCurrentUserName())
	PROFILE_DIRECTORY = "Default"
)

func init() {
	rootCmd.AddCommand(oeiCmd)
	//logrus.Println("")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	oeiCmd.PersistentFlags().BoolP("new", "n", false, "Обновление файлов из раздела Полезные документы")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// oeiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ScrapFilesFromOEIv5() {
	var noticeFmt = color.New(color.FgGreen).PrintlnFunc()
	utils.GetChromePath()
	noticeFmt("[oei-analitika] Обновление...")
	root, _ := os.LookupEnv("FIF_PATH")
	err := utils.EnsureMakeDir(path.Join(root, "OEI"))
	if err != nil {
		logrus.Errorln(err)
	} else {
		noticeFmt("Каталог для файлов создан")
	}

	// create context with opts
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:0], // No default options to provent chrome account login problems.
		chromedp.ExecPath(CHROME_EXEC_PATH),
		chromedp.UserDataDir(USER_DATA_DIR),
		chromedp.Flag("profile-directory", PROFILE_DIRECTORY),
		chromedp.Flag("headless", false),
		chromedp.Flag("flag-switches-begin", true),
		chromedp.Flag("flag-switches-end", true),
		chromedp.Flag("enable-automation", false),
		//chromedp.Flag("disable-blink-features", "AutomationControlled"),
		// chromedp.Flag("new-window", true),
	)

	cr, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel1()

	// ctx, cancel2 := context.WithTimeout(cr, 120*time.Second)
	// defer cancel2()

	// create a new browser
	ctx, cancel3 := chromedp.NewContext(cr, chromedp.WithLogf(logrus.Printf))
	defer cancel3()

	// Получаем таблицу полезных данных
	noticeFmt("Поиск новых документов...")
	res, err := listOEIpages(ctx, "Кто ты, герой?")
	if err != nil {
		logrus.Fatalf("ошибка получения документов: %v", err)
	}

	// Сохраняем таблицу в csv
	saveTableRows(res)

	// Cкачиваем полезные документы
	noticeFmt("Скачиваем...")
	countExist := 0
	for i, rows := range res {
		// skip headers
		if i == 0 {
			continue
		}
		time.Sleep(1000 * time.Millisecond)
		fmt.Println(rows.Link)
		err = utils.DownloadFile(getMPlink(rows.Link), getFName(rows))
		if err != nil {
			if strings.Contains(err.Error(), "file exist") {
				countExist++
			}
			logrus.Error(err)
			continue
		}
		// Если нет новых файлов, выходим
		if countExist > 100 {
			break
		}

	}
	noticeFmt("Обновление завершено успешно!")
}

// listOEIpages is finding the specified section sect, and retrieving the
// associated url from the page OEI.
func listOEIpages(ctx context.Context, sect string) (rows_out []TRows, err error) {

	var cancel func()
	var tableRows, tblrs []TRows

	// Таймаут перелистывания 120 сек,
	// если медленная скорость доступа в интернет, можно увеличить таймаут
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Переходим на сайт
	err = chromedp.Run(ctx,
		emulation.SetUserAgentOverride(utils.UserAgent),
		chromedp.Navigate(`https://oei-analitika.ru/kurilka/reestr_good_docs.php`))
	if err != nil {
		return nil, fmt.Errorf("could not navigate to url: %v", err)
	}

	// Якорь - селектор `того самого`  сайта
	yak := fmt.Sprintf(`//label[text()[contains(., '%s')]]`, sect)

	// Ждем появления Якоря страницы
	err = chromedp.Run(ctx, chromedp.WaitVisible(yak))
	if err != nil {
		return nil, fmt.Errorf("could not get section: %v", err)
	}

	// Получаем число страниц таблицы
	total_p := `#good_docs_paginate > ul > li:nth-child(8) > a`
	var page_n string

	err = chromedp.Run(ctx, chromedp.Text(total_p, &page_n, chromedp.ByQuery))
	if err != nil {
		return nil, fmt.Errorf("could not get pages: %v", err)
	}
	total, _ := strconv.Atoi(page_n)
	//fmt.Println("All pages is", uint16(total))

	var nodes []*cdp.Node
	var rows []*cdp.Node

	// Ищем заголовок таблицы, записываем его в переменную headers
	var headers []*cdp.Node

	err = chromedp.Run(ctx, chromedp.Tasks{chromedp.Nodes(".table-striped.dataTable thead tr th", &headers)})
	if err != nil {
		return nil, fmt.Errorf("could not get table's header: %v", err)
	}

	// Обрабатываем заголовки
	var tableHeader TRows
	for i, elem := range headers {
		switch i {
		case 0:
			tableHeader.Num = strings.Split(elem.Attributes[13], ":")[0]
		case 1:
			tableHeader.DocName = strings.Split(elem.Attributes[11], ":")[0]
		case 2:
			tableHeader.DocDesc = strings.Split(elem.Attributes[11], ":")[0]
		case 3:
			tableHeader.Author = strings.Split(elem.Attributes[11], ":")[0]
		case 4:
			tableHeader.DocType = strings.Split(elem.Attributes[11], ":")[0]
		case 5:
			tableHeader.WhoUp = strings.Split(elem.Attributes[11], ":")[0]
		case 6:
			tableHeader.DateUp = strings.Split(elem.Attributes[11], ":")[0]
		case 7:
			tableHeader.Link = strings.Split(elem.Attributes[11], ":")[0]
		case 8:
			tableHeader.TypeSize = strings.Split(elem.Attributes[11], ":")[0]
		}
	}
	// fmt.Printf("tableHeader is %+v\n", tableHeader)

	// Сортировка от новых к старым
	// #good_docs > thead > tr > th.sorting.sorting_asc
	err = chromedp.Run(ctx, chromedp.Click("#good_docs > thead > tr > th.sorting.sorting_asc"))
	if err != nil {
		return nil, fmt.Errorf("could not get links and descriptions: %v", err)
	}

	// Обходим все страницы
	for i := 2; i <= total-500; i++ {
		var urls []string
		// Ищем на каждой странице ссылки на страницы с методиками
		err = chromedp.Run(ctx, chromedp.Nodes("a", &nodes))
		if err != nil {
			return nil, fmt.Errorf("could not get links and descriptions: %v", err)
		}

		// Заполняем массив urls ссылками с каждой странице
		for _, n := range nodes {
			if strings.Contains(n.AttributeValue("href"), "id_mp=") {
				fmt.Println("mp=", n.AttributeValue("href"))
				urls = append(urls, fmt.Sprintf("https://oei-analitika.ru/kurilka/%s", n.AttributeValue("href")))
			}
		}

		// Ищем все строки таблицы, записываем их в переменную rows
		err := chromedp.Run(ctx, chromedp.Nodes(".table-striped.dataTable > tbody > tr", &rows))
		if err != nil {
			return nil, fmt.Errorf("could not get table's rows: %v", err)
		}
		logrus.Infoln("len rows is", len(rows))

		// Обрабатываем таблицу построчно
		var tableRow TRows
		for j, row := range rows {
			//logrus.Debugln("selected for j row url's", urls[j])
			var text string
			err = chromedp.Run(ctx, chromedp.Text(row.FullXPathByID(), &text))
			if err != nil {
				return nil, fmt.Errorf("could not get row items: %v", err)
			}
			// Разбиваем строки по элементам
			elements := strings.Split(text, "	")

			// Печать элементов строки
			for v, elem := range elements {
				switch v {
				case 0:
					tableRow.Num = elem
				case 1:
					tableRow.DocName = elem
				case 2:
					tableRow.DocDesc = elem
				case 3:
					tableRow.Author = elem
				case 4:
					tableRow.DocType = elem
				case 5:
					tableRow.WhoUp = elem
				case 6:
					tableRow.DateUp = elem
				case 7:
					tableRow.Link = urls[j]
					logrus.Infoln("tableRow.Link is ", j, urls[j])
				case 8:
					tableRow.TypeSize = utils.CleanEOL(elem)
				}
			}
			tableRows = append(tableRows, tableRow)
			// fmt.Printf("tableRow is %+v\n", tableRow)
		}

		// Жмем кнопку перехода на слудеющую страницу
		var next int
		if i <= 6 {
			next = i
		} else {
			next = 6
		}
		err = chromedp.Run(ctx, chromedp.Click(fmt.Sprintf("#good_docs_paginate > ul > li:nth-child(%d) > a", next)))
		if err != nil {
			return nil, fmt.Errorf("could not get links and descriptions: %v", err)
		}
		// Пауза для более плавной отзывчивости сайта
		time.Sleep(50 * time.Millisecond)
	}
	// fmt.Printf("tableRows is %+v\n", tableRows)
	tblrs = prependTRows(tableRows, tableHeader)

	return tblrs, nil
}

func prependTRows(x []TRows, y TRows) []TRows {
	x = append(x, TRows{})
	copy(x[1:], x)
	x[0] = y
	return x
}

// saveTableRows: Save table in csv
func saveTableRows(t []TRows) error {

	// Создание нового CSV-файла для записи данных
	f, err := os.Create("table-data.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	// Создание нового писателя CSV-файла
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Comma = ';'
	defer w.Flush()

	// Запись заголовков в CSV-файл
	// t0 := &t[0]
	// headers := t0.getHeaders(*t0)
	// err = w.Write(headers)
	// if err != nil {
	// 	return err
	// }
	// Запись строк в CSV-файл
	for _, row := range t {
		// if i == 0 {
		// continue
		// }
		err = w.Write([]string{row.DocName, row.DocDesc, row.Author, row.DocType, row.WhoUp, row.DateUp, row.Link, utils.CleanEOL(row.TypeSize)})
		if err != nil {
			return err
		}
	}
	logrus.Println("Table data has been stored to table-data.csv file")
	return nil
}

// getHeaders: return table's headers
func (TRows) getHeaders(location TRows) (header []string) {

	loc := reflect.TypeOf(location)
	if loc.Kind() == reflect.Struct {
		for i := 0; i < loc.NumField(); i++ {
			header = append(header, loc.Field(i).Name)
		}
		return header
	} else {
		fmt.Println("not a stuct")
		return nil
	}
}

func getFName(n TRows) string {
	// if n.DocDesc == "" {
	// 	logrus.Println("Error DocDesk empty")
	// 	//return ""
	// }
	var fn string

	switch n.DocType {
	case "Методика поверки":
		{
			fn = "mp_"

			if len(n.DocDesc) == 0 {
				// fmt.Println("n.Doc", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocName, "/", "_")
			} else {
				// fmt.Println("n.DocName", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocDesc, "/", "_")
			}
		}
	case "Паспорт":
		{
			fn = "passp_"
			if len(n.DocDesc) == 0 {
				// fmt.Println("n.Doc", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocName, "/", "_")
			} else {
				// fmt.Println("n.DocName", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocDesc, "/", "_")
			}
		}
	case "Руководство по эксплуатации":
		{
			fn = "re_"

			if len(n.DocDesc) == 0 {
				// fmt.Println("n.Doc", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocName, "/", "_")
			} else {
				// fmt.Println("n.DocName", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocDesc, "/", "_")
			}
		}
	case "Описание типа":
		{
			fn = "ot_"
			if len(n.DocDesc) == 0 {
				// fmt.Println("n.Doc", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocName, "/", "_")
			} else {
				// fmt.Println("n.DocName", n.DocName)
				fn = fn + strings.ReplaceAll(n.DocDesc, "/", "_")
			}
		}
	}
	// logrus.Debugln(n.DocDesc)

	var ext string
	if strings.Contains(n.TypeSize, "pdf") {
		ext = ".pdf"
	} else if strings.Contains(n.TypeSize, "docx") {
		ext = ".docx"
	} else if strings.Contains(n.TypeSize, "doc") {
		ext = ".doc"
	} else {
		ext = ".filetype"
	}
	// ext := filepath.Ext(n.TypeSize)
	// urls = append(urls, link)

	fullname := path.Join("OEI", strings.ReplaceAll(fmt.Sprintf("%s%s", fn, ext), "/", "_"))
	// strings.Replace(p, ".php?id_mp=", ".pdf", 1)
	// fmt.Println("get fulllname is", fullname)

	return fullname
}

func getMPlink(v string) string {
	// Instantiate default collector
	var out string
	c := colly.NewCollector(
		colly.UserAgent(utils.UserAgent),
	)
	// Настраиваем сетевые параметры и тайм-ауты
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

	// Настраиваем cookie
	c.OnRequest(
		func(r *colly.Request) {
			// r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			// r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
			// r.Headers.Set("Referer", "https://oei-analitika.ru/kurilka/reestr_good_docs.php")
			// r.Headers.Set("Accept-Language", "ru")
			// r.Headers.Set("Connection", "keep-alive")
			// r.Headers.Set("Host", "oei-analitika.ru")
			// r.Headers.Set("Cookie", "PHPSESSID=h5f5o041ihirbquv3u52lcdu3a; _ga_WLPB0SB6JG=GS1.1.1678681504.9.1.1678681554.0.0.0; _ga=GA1.1.1827441606.1676862318; _ym_d=1676862320; _ym_uid=1676862320686846588")
		},
	)
	// Дебаг ответа
	c.OnResponse(func(r *colly.Response) {
		// logrus.Println("Got: ", r.Request.URL)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		if strings.Contains(link, ".pdf") {
			out = fmt.Sprintf("https://oei-analitika.ru/kurilka/%s", link)
			fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		}
	})
	// Обрабатывем ошибки выполнения запроса
	c.OnError(
		func(r *colly.Response, err error) {
			logrus.Error("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		},
	)

	// Start scraping on v
	c.Visit(v)

	return out
}
