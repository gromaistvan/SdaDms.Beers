package main

import (
	"encoding/json"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type resource struct {
	Localizer *i18n.Localizer
}

func loadMessageFiles(bundle *i18n.Bundle, files ...string) {
	for _, file := range files {
		if _, err := bundle.LoadMessageFile(file); err != nil {
			log.Printf("Failed to load message file %s: %v", file, err)
		}
	}
}

func newResource(langugage language.Tag) *resource {
	bundle := i18n.NewBundle(langugage)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	loadMessageFiles(bundle, "en.json", "hu.json")
	return &resource{i18n.NewLocalizer(bundle, langugage.String())}
}

func (res *resource) get(id string) string {
	value, err := res.Localizer.Localize(&i18n.LocalizeConfig{MessageID: id})
	if err != nil {
		log.Printf("Failed to read resource '%s': %v", id, err)
		return "N/A"
	}
	return value
}
