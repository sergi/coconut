package coconut

import (
	"bytes"
	"strconv"
	"text/template"
	"time"
)

func getYear(t time.Time) string {
	return strconv.Itoa(t.Year())
}
func getMonth(t time.Time) string {
	m := t.Month()
	return strconv.Itoa(int(m))
}
func getDay(t time.Time) string {
	return strconv.Itoa(t.Day())
}

func ExecuteConfigTemplate(data Result, pathTemplate string) (string, error) {
	funcMap := template.FuncMap{
		"day":   getDay,
		"month": getMonth,
		"year":  getYear,
	}

	tpl, err := template.New("config").Funcs(funcMap).Parse(pathTemplate)
	if err != nil {
		return "", err
	}
	var tplBytes bytes.Buffer
	err = tpl.Execute(&tplBytes, data)
	if err != nil {
		return "", err
	}
	return tplBytes.String(), nil
}
