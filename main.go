package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/jedib0t/go-pretty/table"
)

type result struct {
	date       string
	aTeam      string
	hTeam      string
	finalScore string
	stadium    string
	network    string
}

type fixtureApiResponse struct {
	Content []apiResult
}

type compSeasonApiResponse struct {
	Content []apiCompetition
}

type apiCompetition struct {
	Id float32
}

type apiResult struct {
	Teams    []apiTeam
	Gameweek apiGameWeek
}

type apiGameWeek struct {
	Gameweek   float32
	CompSeason apiCompSeason
}

type apiCompSeason struct {
	Id float32
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
	gameweek := flag.Int("gameweek", 0, "The gameweek to pull results from.")

	flag.Parse()

	results := make(chan result)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Wait()
		close(results)
	}()

	go scrapeResults(results, gameweek, team, &wg)

	printResults(results)
}

func getJson(url string, target interface{}) error {
	client := http.Client{}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func compSeason() int32 {
	target := compSeasonApiResponse{}

	err := getJson("https://footballapi.pulselive.com/football/competitions/1/compseasons", &target)

	if err != nil {
		log.Fatal(err)
	}

	return int32(target.Content[0].Id)
}

func scrapeResults(results chan<- result, gameweek *int, team *string, wg *sync.WaitGroup) {
	target := fixtureApiResponse{}

	compSeason := compSeason()

	url := fmt.Sprintf("https://footballapi.pulselive.com/football/fixtures?comps=1&pageSize=40&sort=desc&statuses=C&compSeasons=%d", compSeason)

	err := getJson(url, &target)

	if err != nil {
		log.Fatal(err)
	}

	for _, lineitem := range target.Content {
		teams := lineitem.Teams

		gameweekSet := isFlagPassed("gameweek")

		if gameweekSet && int(lineitem.Gameweek.Gameweek) != *gameweek {
			continue
		}

		results <- result{
			hTeam:      teams[0].Team.Name,
			aTeam:      teams[1].Team.Name,
			finalScore: fmt.Sprintf("%d - %d", int(teams[0].Score), int(teams[1].Score)),
		}
	}

	wg.Done()
}

func printResults(results <-chan result) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"HOME", "AWAY", "SCORE"})

	for result := range results {
		t.AppendRow(table.Row{result.hTeam, result.aTeam, result.finalScore})
	}

	t.Render()
}
