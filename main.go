package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bundgaard/js"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"sync"
	"time"

	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

var (
	element    = flag.String("node", "", "element to extract")
	doDownload = flag.Bool("download", false, "download the file")
	url        string
)

func init() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		_, _ = fmt.Fprintf(os.Stderr, "not enough arguments\n")
		os.Exit(1)
	}

	url = flag.Args()[0]
	pattern := regexp.MustCompile(`https?://`)
	if !pattern.MatchString(url) {
		_, _ = fmt.Fprintf(os.Stderr, "wrong URL format\n")
		os.Exit(1)
	}
}

func main() {

	buf := new(bytes.Buffer)
	if err := download(buf, url); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	doc, _ := html.Parse(buf)
	scripts, err := crawlScripts(doc)
	if err != nil {
		log.Fatal(err)
	}

	javascriptFile := new(bytes.Buffer)
	for _, script := range scripts {
		if script.FirstChild != nil {
			if strings.Contains(script.FirstChild.Data, *element) {
				_, _ = fmt.Fprintf(javascriptFile, "%s", script.FirstChild.Data)
			}
		}
	}
	if javascriptFile.Len() < 1 {
		_, _ = fmt.Fprintf(os.Stderr, "could not find %s", *element)
		os.Exit(1)
	}

	_ = ioutil.WriteFile("site.html", javascriptFile.Bytes(), 0600)
	_, environment := js.New(javascriptFile.String())

	fmt.Println(strings.Repeat("=", 80))
	mediaURL, err := environment.GetString("media_0")
	if err != nil {
		log.Fatal("failed to get media_0", err)
	}
	baseURLIdx := strings.LastIndex(mediaURL, "/")
	baseURL := mediaURL[:baseURLIdx+1]
	fmt.Printf("Environment %v\n", environment)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("URL", mediaURL)
	fmt.Println(strings.Repeat("=", 80))

	req, _ := http.NewRequest("GET", mediaURL, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36 Edg/92.0.902.73")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("fish %v", err)
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	ioutil.WriteFile("master.m3u8", content, 0600)

	ch1 := make(chan string)
	ch2 := make(chan string)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		link := <-ch1
		fmt.Printf("Link %s", baseURL+link)
		resp, err := http.Get(baseURL + link)
		if err != nil {
			log.Println(err)
		}
		content, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%d %s %s\n", resp.StatusCode, resp.Status, content)
		if err := readM3U8(content, ch2); err != nil {
			log.Fatal(err)
		}

	}()

	go func() {
		tsFile, err := os.Create("filename.ts")
		if err != nil {
			log.Fatal(err)
		}
		defer tsFile.Close()

		for endpoint := range ch2 {
			downloadTransportStream(tsFile, baseURL, endpoint)
		}

	}()
	if err := readM3U8(content, ch1); err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	type Media struct {
		DefaultQuality bool   `json:"defaultQuality"`
		Format         string `json:"format"`
		Quality        int    `json:"quality,string"`
		VideoURL       string `json:"videoUrl"`
	}

	var mediaList []Media
	if err := json.NewDecoder(resp.Body).Decode(&mediaList); err != nil {
		log.Fatalf("foo %v", err)
	}

	defaultQualityURL := ""
	for _, media := range mediaList {
		if media.DefaultQuality {
			defaultQualityURL = media.VideoURL
			break
		}
	}

	fmt.Println(defaultQualityURL)

	dlinfo, err := httpClient.Head(defaultQualityURL)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "baz %v", err)
		os.Exit(1)
	}

	if dlinfo.StatusCode < 200 || dlinfo.StatusCode >= 400 {
		_, _ = fmt.Fprintf(os.Stderr, "http error %d %s", dlinfo.StatusCode, dlinfo.Status)
		os.Exit(1)
	}
	fmt.Printf("dlinfo %d %s\n", dlinfo.StatusCode, dlinfo.Status)
	contentLength := dlinfo.Header.Get("Content-Length")
	contentType := dlinfo.Header.Get("Content-Type")

	fmt.Printf("Length: %s bytes\n", contentLength)
	fmt.Printf("Type: %s\n", contentType)

	if *doDownload {

		downloadSomething(defaultQualityURL)
	}

}
func downloadSomething(defaultQualityURL string) {
	ctxCancel, cancelFn := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, c := range `-\|/` {
					_, _ = fmt.Fprintf(os.Stdout, "\r%c", c)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}(ctxCancel)
	videoResp, err := httpClient.Get(defaultQualityURL)

	if err != nil {
		log.Fatal(err)
	}
	defer videoResp.Body.Close()

	videoFile, err := os.Create(uuid.NewString() + ".mp4")
	if err != nil {
		log.Fatalf("oscar %v", err)
	}
	defer videoFile.Close()

	_, _ = io.Copy(videoFile, videoResp.Body)
	cancelFn()
}

func downloadTransportStream(tsFile *os.File, baseURL, endpoint string) {
	response, err := http.Get(baseURL + endpoint)
	if err != nil {
		log.Println(err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		log.Printf("%d %s", response.StatusCode, response.Status)
		return
	}

	_, _ = io.Copy(tsFile, response.Body)

}
