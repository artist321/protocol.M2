/*
Copyright © 2023 Artem Demchenko a.a.demchenko@yandex.com
*/
package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"

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

// SetHeader set header for http client
func SetHeader(req *http.Request) {
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

// GetArshinHTTPData get data from site in raw mode
// with setting up (header, transport, timeouts) client
func GetArshinHTTPData(url string) ([]byte, error) {
	log.Debugln(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("error http request init", err)
		return nil, err
	}
	SetHeader(req)
	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		log.Fatalln("error http retry request init from req", err)
		return nil, err
	}
	RetClient.HTTPClient = CustomHTTPClient
	RetClient.RetryMax = 5
	RetClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	RetClient.Logger = nil
	resp, err := RetClient.Do(retryReq)
	if err != nil {
		log.Fatalln("error http request no complete", err)
		return nil, err
	}
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	log.Println("Content-Length: " + resp.Header.Get("Content-Length"))

	if size == 0 {
		log.Debugln("Content-Length: 0")
		//return 2
	}

	if resp.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}(resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("io.ReadAll(resp.Body) ", err, resp.Body)
		return nil, err
	}
	// people1 := people{}
	// jsonErr := json.Unmarshal(body, &people1)
	// if jsonErr != nil {
	//	log.Fatal(jsonErr)
	// }
	return body, nil
}

func CheckMultipart(urls string) (bool, error) {
	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return false, err
	}
	r.Header.Add("Range", "bytes=0-0")
	cl := http.Client{}
	resp, err := cl.Do(r)
	if err != nil {
		log.Errorf(" can't check multipart support assume no %v \n", err)
		return false, err
	}
	if resp.StatusCode != 206 {
		return false, errors.New("file not found or moved status: " + resp.Status)
	}
	if resp.ContentLength == 1 {
		log.Info("multipart download support \n")
		return true, nil
	}
	return false, nil
}

func GetSize(urls string) (int64, error) {
	cl := http.Client{}
	resp, err := cl.Head(urls)
	if err != nil {
		log.Errorf("when try get file size %v \n", err)
		return 0, err
	}
	if resp.StatusCode != 200 {
		log.Error("file not found or moved status:", resp.StatusCode)
		return 0, errors.New("file not found or moved")
	}
	//log.Printf("file size is %d bytes \n", resp.ContentLength)
	return resp.ContentLength, nil
}

func GetChromePath() {

	switch runtime.GOOS {
	case "windows":
		{
			cmd := exec.Command("cmd", "/C", `reg query "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe" /ve | awk '{print $NF}'`)
			out, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
				return
			}
			path := strings.TrimSpace(string(out))
			fmt.Println(path)
		}
	case "darwin":
		{
			cmd := exec.Command(" mdfind ", "'Google Chrome.app' | grep /Applications/ ")
			out, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
				return
			}
			path := strings.TrimSpace(string(out))
			fmt.Println(path)
		}
	case "linux":
		{
			cmd := exec.Command("which", "google-chrome")
			out, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
				return
			}
			path := strings.TrimSpace(string(out))
			fmt.Println(path)
		}
	default:
		{
			fmt.Println("Unsupported OS")
		}
	}
}
