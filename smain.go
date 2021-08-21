package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

func secondMain() {
	buf := new(bytes.Buffer)

	download(buf, "")
	doc, _ := html.Parse(buf)
	scripts, err := crawlScripts(doc)
	if err != nil {
		log.Fatal(err)
	}

	for _, script := range scripts {
		script := script

		if script.FirstChild != nil {

			if strings.Contains(script.FirstChild.Data, "flashvar") {
				f, err := os.Create(filepath.Join("hack", uuid.NewString()+".js"))
				if err != nil {
					log.Println(err)
				}
				fmt.Fprintf(f, "%s", script.FirstChild.Data)
				f.Close()
			}

		}

	}
}

func crawlScripts(doc *html.Node) ([]*html.Node, error) {
	scripts := make([]*html.Node, 0)
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "script" {
			scripts = append(scripts, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}

	crawler(doc)

	if len(scripts) != 0 {
		return scripts, nil
	}
	return nil, fmt.Errorf("didnt find any scripts")
}
