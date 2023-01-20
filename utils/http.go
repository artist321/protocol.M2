/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package utils

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"protocol.M2/log"
	"strconv"
	"strings"
	"time"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"

var RootDir = "."
var RetClient = retryablehttp.NewClient()

var CustomHTTPClient = &http.Client{

	Transport: &http.Transport{
		DisableKeepAlives: true,
		//// If you do have a proxy
		//
		//purl, err := url.Parse("http://444.555.666.777:8888")
		//Proxy: http.ProxyURL(purl)}
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   0 * time.Second, // Connection timeout
			KeepAlive: 30 * time.Second,
		}).DialContext,
		IdleConnTimeout:       0 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 180 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
	},
	Timeout: 0 * time.Second,
}

// DownloadFile скачивает файл по URL в папку
func DownloadFile(url string, fn string) (err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Log.Error("error http request no init", err)

		return errors.New(fmt.Sprintf("download error: http request no init"))
	}
	setHeader(req)
	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		log.Log.Error("error http retry request no init", err)
		return errors.New(fmt.Sprintf("download error: http request no init"))
	}
	RetClient.HTTPClient = CustomHTTPClient
	RetClient.RetryMax = 5
	RetClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	RetClient.Logger = nil
	resp, err := RetClient.Do(retryReq)
	if err != nil {
		log.Log.Error("error http retry request no complete", err)
		return errors.New(fmt.Sprintf("download error: http retry request no complete"))

	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Log.Error("server receive status", resp.Status, fn)
		//err := errors.New(fmt.Sprintf("download error: %s", resp.Status))
		//return err
		// exit if not ok
	}

	// the Header "Content-Length" will let us know
	// the total file size to download
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	if size == 0 {
		log.Log.Error("file size is null", fn)
		//err := errors.New("download error: file size is null")
		//return err
	}

	if log.Log.GetLevel() == logrus.DebugLevel {
		log.Log.Debugln(url)
		log.Log.Debugln("Content-Disposition: " + resp.Header.Get("Content-Disposition"))
		log.Log.Debugln("Content-Length: " + strconv.Itoa(size))
		_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		fn = strings.Replace(params["filename"], "+", "_", -1)
		if fn == "" {
			log.Log.Debugln("DownloadFile: 'filename' has no name")
		} else {
			log.Log.Debugln("DownloadFile: 'filename' name is " + fn)
		}
	}

	name := fn

	bar := progressbar.DefaultBytes(resp.ContentLength, "Cкачиваю '"+name+"' ")
	if resp.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Log.Errorln(fmt.Errorf("%s", err))
			}
		}(resp.Body)
	}
	//
	//body, readErr := ioutil.ReadAll(resp.Body)
	//if readErr != nil {
	//	log.Fatal(readErr)
	//}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Log.Errorln(fmt.Errorf("%s", err))
		}
	}(resp.Body)

	if !IsExist(path.Join(RootDir, fn)) {
		out, err := os.Create(path.Join(RootDir, fn))
		if err != nil {
			return err
		}
		defer func(out *os.File) {
			err := out.Close()
			if err != nil {
				log.Log.Errorln(fmt.Errorf("%s", err))
			}
		}(out)
		_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
		if err != nil {
			log.Log.Errorln(fmt.Errorf("%s", err))
			return err
		}
	} else {
		//if isFileSameSize(path.Join(PathSI, filename), getSize(resp.Body) {
		//	log.Fatal("[E] Такой же файл уже существует")
		//}
		log.Log.Infoln(" Файл уже существует")
	}
	return nil
}

// setHeader set header for http client
func setHeader(req *http.Request) {
	//var t time.Time
	gmtTimeLoc := time.FixedZone("GMT", 0)
	t := time.Date(2016, time.July, 18, 02, 26, 04, 0, gmtTimeLoc)
	//s := t.In(t).Format(http.TimeFormat)
	s := t.Format(http.TimeFormat)
	//log.Println(t.UTC().Format(time.RFC1123))
	//log.Println(s)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set(
		"Cookie",
		"session-cookie=1739c301728ceb5bb9dbe8bc4c95548f905fdc5b4b4d7a236f302e87aba072746255b65b2ddd6607de8666c49903962e; "+
			"REGISTRY_SETTINGS=%5B%5D; JSESSIONID=1E0CCE5FB8C83F7551106BE9BBFC2C75",
	)
	req.Header.Add("Last-Modified", s)
}
