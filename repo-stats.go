package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Configuration contains the config properties.
type Configuration struct {
	Token    string `json:"token"`
	URIRepos string `json:"uri_repos"`
	URIStats string `json:"uri_stats"`
}

// Repository contains the properties for the git repository.
type Repository struct {
	Name           string
	Visibility     string
	Size           int
	Statistics     []Statistic
	TotalCommits   int64
	TotalAdditions int64
	TotalDeletions int64
	NumberAuthors  int
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Sorting!

// By is a less function to define the ordering of the Repository arguments.
type By func(r1, r2 *Repository) bool

// Join the By function and the slice to be sorted.
type repoSorter struct {
	repositories []Repository
	by           func(r1, r2 *Repository) bool // Closure.
}

// Sort is a function for sorting the argument slice.
func (by By) Sort(repositories []Repository) {
	rs := &repoSorter{
		repositories: repositories,
		by:           by,
	}
	sort.Sort(rs)
}

// Len is part of sort.Interface.
func (s *repoSorter) Len() int {
	return len(s.repositories)
}

// Swap is part of sort.Interface.
func (s *repoSorter) Swap(i, j int) {
	s.repositories[i], s.repositories[j] = s.repositories[j], s.repositories[i]
}

// Less is part of sort.Interface. It is implemented by calling the By closure.
func (s *repoSorter) Less(i, j int) bool {
	return s.by(&s.repositories[i], &s.repositories[j])
}

// setConfiguration adds all of the json config settings into the struct.
func setConfiguration(file string) Configuration {
	// Get the config json into the Configuration struct.
	var configuration Configuration
	configJson, err := os.Open(file)
	check(err)
	defer configJson.Close()
	byteValue, _ := ioutil.ReadAll(configJson)
	json.Unmarshal(byteValue, &configuration)
	return configuration
}

// setRequest creates and sends the request.
func setRequest(uri string, token string) []byte {
	// Set up the request.
	client := &http.Client{}
	request, err := http.NewRequest("GET", uri, nil)
	check(err)
	request.Header.Set("Authorization", token)

	// Get the repos json data.
	response, _ := client.Do(request)
	jsonData, _ := ioutil.ReadAll(response.Body)

	return jsonData
}

// getJsonResponse gets the full json object returned from the Request.
func getJsonResponse(uri string, token string, fixer string) []interface{} {
	// Set up the request.
	jsonData := setRequest(uri, token)
	jsonText := string(jsonData)

	// Github returns empty stats data for the first uncached request, so try again.
	if jsonText == "{}" {
		jsonData = setRequest(uri, token)
		jsonText = string(jsonData)
	}

	// Fix the bad json from GitHub.
	jsonText = `{ "` + fixer + `": ` + jsonText + ` }`

	// Unmarshal the json.
	var jsonDataFixed map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDataFixed)

	// Use type assertion to get the repo list into the correct type.
	dataList := jsonDataFixed[fixer].([]interface{})
	return dataList
}

// outputMarkdown sends the repository list to the markdown file.
func outputMarkdown(repositories []Repository) {
	// Create the output file.
	outputFile, err := os.Create("README.md")
	check(err)
	defer outputFile.Close()

	// Add the headers.
	fmt.Fprint(outputFile, "# IAP Repository Stats\n\n")
	fmt.Fprint(outputFile, "The data below is the output of the `repo-stats.go` package.\n\n")
	fmt.Fprint(outputFile, "## All Repositories\n\n")

	fmt.Fprintln(outputFile, "Repository Name | Visibility | Size (kb) | Commits | Additions | Deletions | Authors")
	fmt.Fprintln(outputFile, "--------------- | ---------- | --------- | ------- | --------- | --------- | -------")
	outputFile.Sync()

	// Closures to order the Repository structure.
	size := func(r1, r2 *Repository) bool {
		return r2.Size < r1.Size
	}

	// Sort the repositories by Size.
	By(size).Sort(repositories)

	for i := range repositories {
		fmt.Fprintf(outputFile, "%s | %s | %d | %d | %d | %d | %d\n", repositories[i].Name, repositories[i].Visibility, repositories[i].Size, repositories[i].TotalCommits, repositories[i].TotalAdditions, repositories[i].TotalDeletions, repositories[i].NumberAuthors)
	}
	outputFile.Sync()
}

func main() {
	// Set the local variables.
	configFile := "config/config.json"

	// Get the config json into the Configuration struct.
	configuration := setConfiguration(configFile)

	// Get the json response.
	repoList := getJsonResponse(configuration.URIRepos, configuration.Token, "repos")

	// Declare a slice of all the repos.
	repositories := make([]Repository, len(repoList))

	// Loop through the slice, building the Repository struct.
	for i := range repoList {
		repoName := repoList[i].(map[string]interface{})["name"].(string)
		private := repoList[i].(map[string]interface{})["private"].(bool)

		// Visibility is listed in the markdown as either Public or Private.
		visibility := "public"
		if private {
			visibility = "private"
		}

		size := int(repoList[i].(map[string]interface{})["size"].(float64))

		// For each repo, get the contributor statistics.
		URIStatsItem := strings.Replace(configuration.URIStats, ":repo", repoName, 1)
		statsList := getJsonResponse(URIStatsItem, configuration.Token, "stats")

		// Declare a slice of all the stats
		statistics := make([]Statistic, len(statsList))

		// Declare the totals counters.
		var totalCommits int64
		var totalAdditions int64
		var totalDeletions int64
		var numberAuthors int

		// Loop through the slice, building the Statistics struct.
		for j := range statsList {
			// Type assert the complete stats json object.
			statsItem := statsList[j].(map[string]interface{})

			// Set the total value
			total := int64(statsItem["total"].(float64))

			// Get the "weeks" json object.
			weeksList := statsItem["weeks"].([]interface{})

			// Create a slice for the weeks data.
			weeks := make([]Week, len(weeksList))

			// Loop through the weeks json.
			for k := range weeksList {
				// Set the week items data.
				weekItem := weeksList[k].(map[string]interface{})
				weekNumberUnix := weekItem["w"].(float64)
				weekNumber := time.Unix(int64(weekNumberUnix), 0).Format(time.RFC3339)
				additions := int64(weekItem["a"].(float64))
				deletions := int64(weekItem["d"].(float64))
				commits := int64(weekItem["c"].(float64))

				totalAdditions += additions
				totalDeletions += deletions
				totalCommits += commits

				// Add the values to the Week slice.
				weeks[k] = Week{weekNumber, additions, deletions, commits}
			}

			author := statsItem["author"].(map[string]interface{})
			contributor := author["login"].(string)

			// The json response is sectioned by authors - increment the counter.
			numberAuthors++

			// Add the values to the Statistic slice.
			statistics[j] = Statistic{total, weeks, contributor}
		}

		// Add the values to the Repository struct.
		repositories[i] = Repository{repoName, visibility, size, statistics, totalCommits, totalAdditions, totalDeletions, numberAuthors}
	}

	// Output the repositories in size order.
	outputMarkdown(repositories)
}
