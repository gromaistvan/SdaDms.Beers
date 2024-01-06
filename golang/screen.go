package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"math/rand"
	"net/http"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nfnt/resize"
)

type location int

const (
	up location = iota
	down
	home
	end
)

type screen struct {
	Beers     []beer
	BeerIdx   int
	Resources *resource
	Config    config
	Width     int
	Height    int
}

func newScreen(config config) *screen {
	res, err := http.Get(config.URL)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	var beers []beer
	decoder := json.NewDecoder(res.Body)
	if decoder.Decode(&beers) != nil || len(beers) == 0 {
		return nil
	}

	width, height := ui.TerminalDimensions()

	return &screen{
		Beers:     beers,
		BeerIdx:   -1,
		Config:    config,
		Width:     width,
		Height:    height,
		Resources: newResource(config.Language)}
}

func (sc screen) beer() *beer {
	return &sc.Beers[sc.BeerIdx]
}

func (sc *screen) next() bool {
	if sc.BeerIdx < len(sc.Beers)-1 {
		sc.BeerIdx++
		return true
	}
	return false
}

func (sc *screen) prev() bool {
	if sc.BeerIdx > 0 {
		sc.BeerIdx--
		return true
	}
	return false
}

func (sc *screen) first() bool {
	if sc.BeerIdx != 0 {
		sc.BeerIdx = 0
		return true
	}
	return false
}

func (sc *screen) last() bool {
	if sc.BeerIdx != len(sc.Beers)-1 {
		sc.BeerIdx = len(sc.Beers) - 1
		return true
	}
	return false
}

func (sc screen) colorize(block *ui.Block) {
	block.TitleStyle.Bg = sc.Config.Background
	block.TitleStyle.Fg = ui.ColorYellow
	block.BorderStyle.Bg = sc.Config.Background
	block.BorderStyle.Fg = ui.ColorYellow
}

func (sc screen) drawList() ui.Drawable {
	list := widgets.NewList()
	list.Title = sc.Resources.get("list.title")
	list.Rows = make([]string, len(sc.Beers))
	maxLength := 0
	for i := range sc.Beers {
		text := fmt.Sprintf("%s (%d)", sc.Beers[i].Name, sc.Beers[i].ID)
		if maxLength < len(text) {
			maxLength = len(text)
		}
		list.Rows[i] = text
	}
	list.SelectedRow = sc.BeerIdx
	list.SetRect(0, 0, maxLength+1+2, sc.Height)
	sc.colorize(&list.Block)
	list.SelectedRowStyle.Fg = ui.ColorYellow
	return list
}

func (sc screen) drawName(from, to int) ui.Drawable {
	widget := widgets.NewParagraph()
	widget.Title = sc.Resources.get("name.title")
	widget.Text = sc.beer().Name
	widget.WrapText = true
	widget.SetRect(from, 0, to, 3)
	sc.colorize(&widget.Block)
	return widget
}

func (sc screen) drawTagline(from, to int) ui.Drawable {
	widget := widgets.NewParagraph()
	widget.Title = sc.Resources.get("tagline.title")
	widget.Text = sc.beer().Tagline
	widget.WrapText = true
	widget.SetRect(from, 3, to, 6)
	sc.colorize(&widget.Block)
	return widget
}

func (sc screen) drawDescription(from, to int) ui.Drawable {
	widget := widgets.NewParagraph()
	widget.Title = sc.Resources.get("description.title")
	widget.Text = sc.beer().Description
	widget.WrapText = true
	widget.SetRect(from, 6, to, 6+(sc.Height-6-3)/2)
	sc.colorize(&widget.Block)
	return widget
}

func (sc screen) drawIbu(from, to int) ui.Drawable {
	const maxIbu = 150
	value := int(sc.beer().Ibu)
	widget := widgets.NewGauge()
	widget.Title = sc.Resources.get("ibu.title")
	widget.Percent = value * 100 / maxIbu
	widget.Label = fmt.Sprintf("%d %s", value, sc.Resources.get("ibu.unit"))
	widget.SetRect(from, 6+(sc.Height-6-3)/2, to, 9+(sc.Height-6-3)/2)
	sc.colorize(&widget.Block)
	widget.BarColor = ui.ColorYellow
	widget.LabelStyle.Fg = ui.ColorYellow
	return widget
}

func (sc screen) drawIngredients(from, to int) ui.Drawable {
	widget := widgets.NewTable()
	widget.Rows = make([][]string, len(sc.beer().Ingredients.Malt)+1)
	widget.Rows[0] = []string{
		sc.Resources.get("ingredients.malt"),
		sc.Resources.get("ingredients.amount"),
		sc.Resources.get("ingredients.unit")}
	for i, malt := range sc.beer().Ingredients.Malt {
		widget.Rows[i+1] = []string{
			malt.Name,
			fmt.Sprintf("%f", malt.Amount.Value),
			malt.Amount.Unit}
	}
	widget.Title = sc.Resources.get("ingredients.title")
	widget.RowSeparator = true
	widget.FillRow = true
	widget.SetRect(from, 9+(sc.Height-6-3)/2, to, sc.Height)
	sc.colorize(&widget.Block)
	return widget
}

func (sc screen) drawImage() ui.Drawable {
	widget := widgets.NewImage(nil)
	widget.Title = sc.Resources.get("image.title")
	resp, err := http.Get(sc.beer().ImageURL)
	if err != nil {
		return widget
	}
	image, _, err := image.Decode(resp.Body)
	if err != nil {
		return widget
	}
	widget.Image = resize.Thumbnail(uint((sc.Width-2)/2), uint(sc.Height-2), image, resize.InterpolationFunction(rand.Intn(5)))
	widget.SetRect(sc.Width-widget.Image.Bounds().Dx()-2, 0, sc.Width, sc.Height)
	sc.colorize(&widget.Block)
	return widget
}

func (sc *screen) draw(location location) error {
	switch location {
	case up:
		sc.prev()
	case down:
		sc.next()
	case home:
		sc.first()
	case end:
		sc.last()
	}
	if sc.BeerIdx <= -1 || len(sc.Beers) <= sc.BeerIdx {
		return fmt.Errorf("BeerIdx (%d) is out of bounds [0, %d]", sc.BeerIdx, len(sc.Beers)-1)
	}
	sc.Width, sc.Height = ui.TerminalDimensions()
	list := sc.drawList()
	image := sc.drawImage()
	left, right := list.GetRect().Max.X, image.GetRect().Min.X
	ui.Render(
		list,
		sc.drawName(left, right),
		sc.drawTagline(left, right),
		sc.drawDescription(left, right),
		sc.drawIbu(left, right),
		sc.drawIngredients(left, right),
		image)
	return nil
}
