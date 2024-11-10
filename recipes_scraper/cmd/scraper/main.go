package main

import (
	// "io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Sprootsy/recipes_scraper/html"
)


func doRequest(req *http.Request) *http.Response {
	resp, errResp := http.DefaultClient.Do(req)
	if errResp != nil {
		log.Println(errResp)
		os.Exit(1)
	}

	if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		log.Println("Too many options", resp)
		os.Exit(2)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Println("Request was not accepted. Status code:", resp.StatusCode)
	}

	return resp
}


func main() {
	req, errReq := http.NewRequest(http.MethodGet, "https://www.giallozafferano.it/ricerca-ricette/torta+salata+pollo+funghi/", nil)
	if errReq != nil {
		log.Println(errReq)
		os.Exit(1)
	}
	resp := doRequest(req)
	if resp.StatusCode >= http.StatusMultipleChoices {
		log.Fatalln("status code", resp.StatusCode)
	}
	
	tokens, errParse := html.Parse(resp.Body)
	if errParse != nil {
		log.Fatalln(errParse)
	}
	rsltOut := strings.Builder{}
	for _, t := range tokens {
		rsltOut.WriteString(t.String())
		rsltOut.WriteRune('\n')
	}
	log.Println(rsltOut.String())
}
