package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type beer struct {
	ID          int
	Name        string
	Tagline     string
	Description string
	ImageURL    string `json:"image_url"`
	Ibu         float32
	Ingredients struct {
		Malt []struct {
			Name   string
			Amount struct {
				Value float32
				Unit  string
			}
		}
	}
}

type beers struct {
	Beers        *[]beer
	SelectedBeer *beer
}

var database []beer

var indexPage *template.Template

func main() {
	res, err := http.Get("https://api.punkapi.com/v2/beers")
	if err != nil {
		log.Fatal("download error: ", err)
	}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&database); err != nil {
		log.Fatal("decode error: ", err)
	}
	res.Body.Close()
	indexPage, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal("template parsing error: ", err)
	}
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("", err)
	}
}

func notFound(w http.ResponseWriter, err error) {
	if err != nil {
		log.Print(err)
	}
	http.Error(w, "404 not found.", http.StatusNotFound)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFound(w, nil)
		return
	}
	switch r.Method {
	case "GET":
		indexPage.Execute(w, beers{&database, &database[0]})
	case "POST":
		if err := r.ParseForm(); err != nil {
			notFound(w, err)
			return
		}
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			notFound(w, err)
			return
		}
		indexPage.Execute(w, beers{&database, &database[id-1]})
	default:
		notFound(w, nil)
	}
}
