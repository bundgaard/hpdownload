package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"regexp"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bundgaard/js"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

var (
	element    = flag.String("node", "", "element to extract")
	doDownload = flag.Bool("download", false, "download the file")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "not enough arguments\n")
		os.Exit(1)
	}

	url := flag.Args()[0]
	pattern := regexp.MustCompile(`https?://`)
	if !pattern.MatchString(url) {
		fmt.Fprintf(os.Stderr, "wrong URL format\n")
		os.Exit(1)
	}

	buf := new(bytes.Buffer)
	if err := download(buf, url); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
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
				fmt.Fprintf(javascriptFile, "%s", script.FirstChild.Data)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 80))

	fmt.Println(javascriptFile.String())
	fmt.Println(strings.Repeat("=", 80))

	if javascriptFile.Len() < 1 {
		fmt.Fprintf(os.Stderr, "could not find %s", *element)
		os.Exit(1)
	}
	scanner := js.NewScanner(javascriptFile.String())
	parser := js.NewParser(scanner)
	program := parser.Parse()

	environment := make(map[string]js.Object)
	js.Eval(program, environment)

	fmt.Println(strings.Repeat("=", 80))
	mediaURL := environment["media_0"].(*js.StringObject).Value
	fmt.Println("URL", mediaURL)
	fmt.Println(strings.Repeat("=", 80))
	req, _ := http.NewRequest("GET", mediaURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36 Edg/92.0.902.73")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	type Media struct {
		DefaultQuality bool   `json:"defaultQuality"`
		Format         string `json:"format"`
		Quality        int    `json:"quality,string"`
		VideoURL       string `json:"videoUrl"`
	}

	var mediaList []Media
	if err := json.NewDecoder(resp.Body).Decode(&mediaList); err != nil {
		log.Fatal(err)
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
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	if dlinfo.StatusCode < 200 || dlinfo.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "http error %d %s", dlinfo.StatusCode, dlinfo.Status)
		os.Exit(1)
	}
	fmt.Printf("dlinfo %d %s\n", dlinfo.StatusCode, dlinfo.Status)
	contentLength := dlinfo.Header.Get("Content-Length")
	contentType := dlinfo.Header.Get("Content-Type")

	fmt.Printf("Length: %s bytes\n", contentLength)
	fmt.Printf("Type: %s\n", contentType)

	if *doDownload {

		ctxCancel, cancelFn := context.WithCancel(context.Background())

		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					for _, c := range `-\|/` {
						fmt.Fprintf(os.Stdout, "\r%c", c)
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}(ctxCancel)
		videoResp, err := httpClient.Get(defaultQualityURL)

		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		videoFile, err := os.Create(uuid.NewString() + ".mp4")
		if err != nil {
			log.Fatal(err)
		}
		defer videoFile.Close()

		io.Copy(videoFile, videoResp.Body)
		cancelFn()
	}

}
