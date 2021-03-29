package main

import (
	"fmt"
	log "github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

type Item struct {
	Ref, Image, Title string
}

type Article struct {
	Title, Text string
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isClass(node *html.Node, tag string, class string) bool {
	return isElem(node, tag) && getAttr(node, "class") == class
}

func parseURL(unprsd string) string {
	return strings.Split(strings.Split(unprsd, "(")[1], ")")[0]
}

func readItem(item *html.Node) *Item {
	if a := item.FirstChild.FirstChild; isElem(a, "a") {
		if cs := getChildren(a); len(cs) == 2 &&
			isClass(cs[0], "div", "photo__inner") &&
			isClass(cs[1], "span", "photo__captions") {
			return &Item{
				Ref:   "https://news.mail.ru" + getAttr(a, "href"),
				Image: parseURL(getAttr(cs[0].FirstChild, "style")),
				Title: cs[1].FirstChild.FirstChild.Data,
			}
		}
	}
	return nil
}

func downloadNews(address string) []*Item {
	log.Info("sending request to news.mail.ru")
	if response, err := http.Get(address); err != nil {
		log.Error("request to news.mail.ru/ failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from news.mail.ru", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from news.mail.ru", "error", err)
			} else {
				return search(doc)
			}
		}
	}
	return nil
}

func search(node *html.Node) []*Item {
	if isClass(node, "div", "grid__row grid__row_height_240") {
		var items []*Item
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if isClass(c, "div", "grid__item grid__item_small_percent-50 grid__item_medium_percent-50 grid__item_large_percent-50") {
				if item := readItem(c); item != nil {
					items = append(items, item)
				}
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c); items != nil {
			return items
		}
	}
	return nil
}

func downloadArticle(address string) *Article {
	log.Info("sending request to news.mail.ru")
	if response, err := http.Get(address); err != nil {
		log.Error("request to news.mail.ru/ failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from news.mail.ru", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from news.mail.ru", "error", err)
			} else {
				return scan_article(doc)
			}
		}
	}
	return nil
}

func scan_article(node *html.Node) *Article {
	if isClass(node, "div", "article__text js-module js-view js-mediator-article js-smoky-links") {
		var article = &Article{
			Title: "",
			Text: "",
		}
		for item := node.FirstChild; item != nil; item = item.NextSibling {
			if isClass(item, "div", "article__item article__item_alignment_left article__item_html") {
				if p := item.FirstChild; isElem(p, "p"){
					for piece := p.FirstChild; piece != nil; piece = piece.NextSibling {
						if isText(piece) {
							article.Text += piece.Data
						} else if isElem(piece, "nobr") && isText(piece.FirstChild){
							article.Text += piece.FirstChild.Data
						}
					}
				}
			}
		}
		return article
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if article := scan_article(c); article != nil {
			return article
		}
	}
	return nil
}

func main() {
	log.Info("Download started")
	items := downloadNews("https://news.mail.ru/incident/")
	for n, i := range items {
		fmt.Printf("[%d] %s\n	Фото: %s\n", n+1, i.Title, i.Ref)
	}
}
