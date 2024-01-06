package main

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
