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
