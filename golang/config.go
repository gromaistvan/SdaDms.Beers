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
	urlFlag := flag.String("url", "https://api.punkapi.com/v2/beers", "The URL to fetch.")
	bgFlag := flag.String("background", "", "The background color.")
	langFlag := flag.String("language", "en", "UI langugage.")
	flag.Parse()
	var background ui.Color
	switch *bgFlag {
	case "black":
		background = ui.ColorBlack
	case "red":
		background = ui.ColorRed
	case "green":
		background = ui.ColorGreen
	case "yellow":
		background = ui.ColorYellow
	case "blue":
		background = ui.ColorBlue
	case "magenta":
		background = ui.ColorMagenta
	case "cyan":
		background = ui.ColorCyan
	case "white":
		background = ui.ColorWhite
	default:
		background = ui.ColorClear
	}
	lang, err := language.Parse(*langFlag)
	if err != nil {
		log.Printf("Error parsing language flag: %v. Defaulting to English.", err)
		lang = language.English
	}
	return &config{URL: *urlFlag, Background: background, Language: lang}
}
