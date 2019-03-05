package main

import (
	"strings"
	"time"
)

// Repository contains the properties for the git repository.
type Repository struct {
	Name       string
	Visibility string
	Size       int
	Statistics []Statistic
	TotalStats Totals
}

// Statistic contains the properties for the total repository statistics.
type Statistic struct {
	Total  int64
	Weeks  []Week
	Author string
}

// Week contains the properties for weekly consolidated data.
type Week struct {
	WeekNumber string
	Additions  int64
	Deletions  int64
	Commits    int64
}

// Totals contains the total counters for the repository statistics.
type Totals struct {
	Commits   int64
	Additions int64
	Deletions int64
	Authors   int
}

// Adds values to the Repository slice.
func (r *Repository) addRepo(repo map[string]interface{}, statistics []Statistic, totals Totals) {
	r.Name = repo["name"].(string)
	r.Visibility = checkVisibility(repo["private"].(bool))
	r.Size = int(repo["size"].(float64))
	r.Statistics = statistics
	r.TotalStats = totals
}

// Add the values to the Statistic slice.
func (s *Statistic) addStats(statsItem map[string]interface{}, weeks []Week) {
	s.Total = int64(statsItem["total"].(float64))
	s.Weeks = weeks
	author := statsItem["author"].(map[string]interface{})
	s.Author = author["login"].(string)
}

// Add the values to the Week slice.
func (w *Week) addWeek(weekItem map[string]interface{}) {
	weekNumberUnix := weekItem["w"].(float64)
	w.WeekNumber = time.Unix(int64(weekNumberUnix), 0).Format(time.RFC3339)
	w.Additions = int64(weekItem["a"].(float64))
	w.Deletions = int64(weekItem["d"].(float64))
	w.Commits = int64(weekItem["c"].(float64))
}

// Increase the Repository totals.
func (t *Totals) increaseActivity(weekItem map[string]interface{}) {
	t.Additions += int64(weekItem["a"].(float64))
	t.Deletions += int64(weekItem["d"].(float64))
	t.Commits += int64(weekItem["c"].(float64))
}

// Increment the Repository total authors.
func (t *Totals) incrementAuthors() {
	t.Authors++
}

// Visibility is listed in the markdown as either Public or Private.
func checkVisibility(private bool) string {
	if private {
		return "private"
	}
	return "public"
}

func repoStatsURI(uri string, repoName string) string {
	return strings.Replace(uri, ":repo", repoName, 1)
}

// Error check.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Get the config json into the Configuration struct.
	configuration := setConfiguration()

	// Get the json response.
	repoList := getJsonResponse(configuration.URIRepos, configuration.Token, "repos")

	// Declare a slice of all the repos.
	repositories := make([]Repository, len(repoList))

	// Loop through the slice, building the Repository struct.
	for i := range repoList {
		// Only get stats for non-forked repos.
		if repoList[i].(map[string]interface{})["fork"].(bool) {
			continue
		}

		// For each repo, get the contributor statistics.
		uri := repoStatsURI(configuration.URIStats, repoList[i].(map[string]interface{})["name"].(string))
		statsList := getJsonResponse(uri, configuration.Token, "stats")

		// Declare a slice of all the stats
		statistics := make([]Statistic, len(statsList))

		// Declare the totals counters.
		var totals Totals

		// Loop through the slice, building the Statistics struct.
		for j := range statsList {
			// Type assert the complete stats json object.
			statsItem := statsList[j].(map[string]interface{})

			// Get the "weeks" json object.
			weeksList := statsItem["weeks"].([]interface{})

			// Create a slice for the weeks data.
			weeks := make([]Week, len(weeksList))

			// Loop through the weeks json.
			for k := range weeksList {
				// Add the values to the Week slice.
				weeks[k].addWeek(weeksList[k].(map[string]interface{}))

				// Add the weekly totals to the repository Totals struct.
				totals.increaseActivity(weeksList[k].(map[string]interface{}))
			}

			// The json response is sectioned by authors - increment the counter.
			totals.incrementAuthors()

			// Add the values to the Statistic slice.
			statistics[j].addStats(statsItem, weeks)
		}

		// Add the values to the Repository slice.
		repositories[i].addRepo(repoList[i].(map[string]interface{}), statistics, totals)
	}

	// Output the repositories in size order.
	outputMarkdown(repositories)
}
