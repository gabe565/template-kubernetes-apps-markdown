package cmd

import (
	"html/template"
	"strconv"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"rowspan": rowspan,
	}
}

func rowspan(n int) template.HTMLAttr {
	if n == 1 {
		return ""
	}
	return template.HTMLAttr(` rowspan="` + strconv.Itoa(n) + `"`)
}
