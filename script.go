package main

import (
	"fmt"

	"golang.org/x/net/html"
)

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
