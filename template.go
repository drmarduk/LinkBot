package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func (t *Template) Load(file string) error {

	data, err := ioutil.ReadFile("html/" + file)
	if err != nil {
		log.Print(err.Error())
		return err
	}

	t.Content = string(data)
	return nil
}

func (t *Template) SetValue(pattern, value string) {
	t.Content = strings.Replace(t.Content, pattern, value, -1)
}

func (t *Template) String() string {
	return t.Content
}
