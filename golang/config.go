package main

import (
	"flag"
	"log"

	ui "github.com/gizak/termui/v3"
	"golang.org/x/text/language"
)

type config struct {
	URL        string
	Background ui.Color
	Language   language.Tag
}

func loadConfig() *config {
	var url string
	flag.StringVar(&url, "url", "https://api.punkapi.com/v2/beers", "The URL to fetch")
	flag.StringVar(&url, "u", "https://api.punkapi.com/v2/beers", "The URL to fetch (sorthand)")
	var background string
	flag.StringVar(&background, "background", "", "The background color")
	flag.StringVar(&background, "b", "", "The background color (shorthand)")
	var languageName string
	flag.StringVar(&languageName, "language", "en", "UI langugage")
	flag.StringVar(&languageName, "l", "en", "UI langugage (shorthand)")
	flag.Parse()
	var backgroundColor ui.Color
	switch background {
	case "black":
		backgroundColor = ui.ColorBlack
	case "red":
		backgroundColor = ui.ColorRed
	case "green":
		backgroundColor = ui.ColorGreen
	case "yellow":
		backgroundColor = ui.ColorYellow
	case "blue":
		backgroundColor = ui.ColorBlue
	case "magenta":
		backgroundColor = ui.ColorMagenta
	case "cyan":
		backgroundColor = ui.ColorCyan
	case "white":
		backgroundColor = ui.ColorWhite
	default:
		backgroundColor = ui.ColorClear
	}
	tag, err := language.Parse(languageName)
	if err != nil {
		log.Printf("Error parsing language flag: %v. Defaulting to English.", err)
		tag = language.English
	}
	return &config{URL: url, Background: backgroundColor, Language: tag}
}
