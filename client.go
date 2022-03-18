package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
)

var httpClient http.Client

type Client struct {
	Jar http.CookieJar
}

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	httpClient = http.Client{
		Jar: jar,
	}
}

func download(w io.Writer, url string) error {

	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("download %d %s", resp.StatusCode, resp.Status)
	}

	for _, c := range resp.Cookies() {
		fmt.Println(c)
	}

	_, _ = io.Copy(w, resp.Body)
	defer resp.Body.Close()
	return nil

}
