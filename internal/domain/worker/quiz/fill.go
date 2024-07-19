package quiz

import (
	"bytes"
	"golang.org/x/net/html"
)

func parseAndFillForm(htmlContent []byte) (map[string]string, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	traverseAndFillForm(doc, data)

	return data, nil
}

// Функция для обхода узлов HTML и вызова соответствующих функций
func traverseAndFillForm(n *html.Node, data map[string]string) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "input":
			processInput(n, data)
		case "select":
			processSelect(n, data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverseAndFillForm(c, data)
	}
}

// Функция для обработки тегов <input> и заполнения текстовых полей
func processInput(n *html.Node, data map[string]string) {
	var name, inputType, value string
	for _, attr := range n.Attr {
		switch attr.Key {
		case "name":
			name = attr.Val
		case "type":
			inputType = attr.Val
		case "value":
			value = attr.Val
		}
	}

	switch inputType {
	case "text":
		if name != "" {
			data[name] = "test"
		}
	case "radio":
		if name != "" {
			if existingValue, exists := data[name]; !exists || len(value) > len(existingValue) {
				data[name] = value
			}
		}
	}
}

// Функция для обработки тегов <select> и выбора опции с самым длинным значением
func processSelect(n *html.Node, data map[string]string) {
	var name string
	var options []string
	for _, attr := range n.Attr {
		if attr.Key == "name" {
			name = attr.Val
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "option" {
			for _, attr := range c.Attr {
				if attr.Key == "value" {
					options = append(options, attr.Val)
				}
			}
		}
	}
	if name != "" && len(options) > 0 {
		data[name] = findLongestValue(options)
	}
}
func findLongestValue(options []string) string {
	longest := ""
	for _, option := range options {
		if len(option) > len(longest) {
			longest = option
		}
	}
	return longest
}
