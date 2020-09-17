package main

import (
	"bytes"
	"text/template"
)

func loadTemplate(path string) (*template.Template, error) {
	t, err := template.ParseFiles(path)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func executeTemplate(t *template.Template, data map[string]interface{}) (bytes.Buffer, error) {
	data["CoswinUsername"] = CoswinUsername
	data["CoswinPassword"] = CoswinPassword
	data["CoswinDatasource"] = CoswinDatasource

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return buf, err
	}

	return buf, nil
}
