package main

import (
	"log"

	ui "github.com/gizak/termui/v3"
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize termui: %v", err)
	}
	defer ui.Close()

	config := loadConfig()
	screen := newScreen(*config)
	err := screen.draw(home)
	if err != nil {
		log.Printf("Failed to draw screen: %v", err)
		return
	}
	for e := range ui.PollEvents() {
		switch e.ID {
		case "<Escape>", "<C-c>":
			return
		case "<Down>":
			err = screen.draw(down)
		case "<Up>":
			err = screen.draw(up)
		case "<Home>":
			err = screen.draw(home)
		case "<End>":
			err = screen.draw(end)
		}
		if err != nil {
			log.Printf("Failed to draw screen: %v", err)
			return
		}
	}
}
