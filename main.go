package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type PageData struct {
	Results string
}

func scrapeWebsite(targetURL string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	var headlines []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(`.titleline`, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll(".titleline")).map(e => e.innerText)`, &headlines),
	)

	if err != nil {
		return "", err
	}

	// writing in text file
	resultText := strings.Join(headlines, "\n")
	os.WriteFile("results.txt", []byte(resultText), 0644)

	return resultText, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// in page load
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		// in Button Click
		url := r.FormValue("url")
		fmt.Println("Scraping:", url)

		results, err := scrapeWebsite(url)
		if err != nil {
			http.Error(w, "Scraping failed: "+err.Error(), 500)
			return
		}

		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, PageData{Results: results})
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/scrape", handler)

	fmt.Println("Server Started: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
