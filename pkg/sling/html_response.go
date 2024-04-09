package sling

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"strings"
)

type HTMLData struct {
	Status         int
	Message        string
	Location       string
	ParentLocation string
	Path           string
	Referer        string
	ChangeLog      string
}

func HtmlData(html string) (data HTMLData, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return data, err
	}

	data.Status = cast.ToInt(htmlElementText(doc, "#Status", "0"))
	data.Message = htmlElementText(doc, "#Message", "Unknown error")
	data.Location = htmlElementHref(doc, "#Location", "")
	data.ParentLocation = htmlElementHref(doc, "#ParentLocation", "")
	data.Path = htmlElementText(doc, "#Path", "")
	data.Referer = htmlElementHref(doc, "#Referer", "")
	data.ChangeLog = htmlElementText(doc, "#ChangeLog", "")

	return data, nil
}

func (d HTMLData) IsError() bool {
	return d.Status <= 0 || d.Status > 399
}

func htmlElementText(doc *goquery.Document, selector string, defaultValue string) string {
	selection := doc.Find(selector)
	if len(selection.Nodes) > 0 {
		return selection.Text()
	}
	return defaultValue
}

func htmlElementHref(doc *goquery.Document, selector string, defaultValue string) string {
	selection := doc.Find(selector)
	if len(selection.Nodes) > 0 {
		href, _ := selection.Attr("href")
		return href
	}
	return defaultValue
}
