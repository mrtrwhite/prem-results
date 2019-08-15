package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/jinzhu/now"
)

type result struct {
	date       string
	aTeam      string
	hTeam      string
	finalScore string
	stadium    string
	network    string
}

type apiResponse struct {
	Content []apiResult
}

type apiResult struct {
	Teams []apiTeam
}

type apiTeam struct {
	Score float32
	Team  apiTeamDetail
}

type apiTeamDetail struct {
	Name string
}

func main() {
	team := flag.String("team", "", "The team you want results for.")
	start := flag.String("start", now.BeginningOfWeek().String(), "The start of the gameweek.")
	end := flag.String("end", now.EndOfWeek().String(), "The end of the gameweek.")

	flag.Parse()

	results := make(chan result)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Wait()
		close(results)
	}()

	go scrapeResults(results, start, end, team, &wg)

	for result := range results {
		fmt.Printf("%s vs %s: %s", result.hTeam, result.aTeam, result.finalScore)
		fmt.Println()
	}
}

func getJson(url string, target *apiResponse) error {
	client := http.Client{}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func scrapeResults(results chan<- result, start, end, team *string, wg *sync.WaitGroup) {
	target := apiResponse{}

	err := getJson("https://footballapi.pulselive.com/football/fixtures?comps=1&pageSize=40&sort=desc&statuses=C", &target)

	if err != nil {
		log.Fatal(err)
	}

	for _, lineitem := range target.Content {
		teams := lineitem.Teams

		results <- result{
			hTeam:      teams[0].Team.Name,
			aTeam:      teams[1].Team.Name,
			finalScore: fmt.Sprintf("%d - %d", int(teams[0].Score), int(teams[1].Score)),
		}
	}

	wg.Done()
}
