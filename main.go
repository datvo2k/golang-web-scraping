package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type Context struct {
	Title   string
	URL     string
	Price   string
	Area    string
	Contact string
	// Location string
}

func main() {
	fakeChrome := req.DefaultClient().ImpersonateChrome()
	c := colly.NewCollector(colly.MaxDepth(2),
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")))
	c.SetClient(&http.Client{Transport: fakeChrome.Transport})
	c.SetRequestTimeout(120 * time.Second)
	contexts := make([]Context, 0)

	//Callbacks
	c.OnHTML("div.content-items", func(e *colly.HTMLElement) {
		e.ForEach("div.content-item", func(i int, h *colly.HTMLElement) {
			item := Context{}
			item.Title = h.ChildText("div.ct_title")
			item.URL = "https://i-batdongsan.com/" + h.ChildAttr("a[href]", "href")
			item.Price = h.ChildText("div.ct_price")
			item.Area = h.ChildText("div.ct_dt")
			item.Contact = h.ChildText("div.ct_contact")
			contexts = append(contexts, item)
		})
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Got this error:", e)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		js, err := json.MarshalIndent(contexts, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Writing data to file")
		if err := os.WriteFile("contexts.json", js, 0664); err == nil {
			fmt.Println("Data written to file successfully")
		}

	})

	c.Visit("https://i-batdongsan.com/can-ban-nha-dat.htm")
}
