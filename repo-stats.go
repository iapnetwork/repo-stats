package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Configuration struct {
	Token string `json:"token"`
}

type Repository struct {
	Name           string
	Private        bool
	Size           float64
	Statistics     []*Statistic
	TotalCommits   float64
	TotalAdditions float64
	TotalDeletions float64
	NumberAuthors  int
}

type Statistic struct {
	Total  float64
	Weeks  []*Week
	Author string
}

type Week struct {
	WeekNumber string
	Additions  float64
	Deletions  float64
	Commits    float64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func SetConfiguration(file string) Configuration {
	// Get the config json into the Configuration struct.
	var configuration Configuration
	configJson, err := os.Open(file)
	check(err)
	defer configJson.Close()
	byteValue, _ := ioutil.ReadAll(configJson)
	json.Unmarshal(byteValue, &configuration)
	return configuration
}

func SetRequest(uri string, token string) []byte {
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

func GetJsonResponse(uri string, token string, fixer string) []interface{} {
	// Set up the request.
	jsonData := SetRequest(uri, token)
	jsonText := string(jsonData)

	// Github sometimes returns empty stats data for the repo, so try again.
	if jsonText == "{}" {
		jsonData = SetRequest(uri, token)
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

func main() {
	// Set the local variables.
	configFile := "config/config.development.json"
	uriRepo := "https://api.github.com/orgs/iapnetwork/repos"
	uriStats := "https://api.github.com/repos/iapnetwork/:repo/stats/contributors"

	// Get the config json into the Configuration struct.
	configuration := SetConfiguration(configFile)

	// Create the output file.
	outputFile, err := os.Create("README.md")
	check(err)
	defer outputFile.Close()

	// Add the headers.
	fmt.Fprint(outputFile, "# IAP Repository Stats\n\n")
	fmt.Fprint(outputFile, "The data below is the output of the `github-stats.go` package.\n\n")
	fmt.Fprint(outputFile, "## All Repositories\n\n")

	fmt.Fprintln(outputFile, "Repository | Private | Size | Commits | Additions | Deletions | Authors")
	fmt.Fprintln(outputFile, "---------- | ------- | ---- | ------- | --------- | --------- | -------")
	outputFile.Sync()

	// Get the json response.
	repoList := GetJsonResponse(uriRepo, configuration.Token, "repos")

	// Declare a slice of all the repos.
	repositories := make([]*Repository, len(repoList))

	// Loop through the slice, building the Repository struct.
	for i := range repoList {
		repoName := repoList[i].(map[string]interface{})["name"].(string)

		// Ignore repos marked for delete.
		if strings.Contains(repoName, "_delete") {
			continue
		}

		private := repoList[i].(map[string]interface{})["private"].(bool)
		size := repoList[i].(map[string]interface{})["size"].(float64)

		// For each repo, get the contributor statistics.
		uriStatsItem := strings.Replace(uriStats, ":repo", repoName, 1)
		statsList := GetJsonResponse(uriStatsItem, configuration.Token, "stats")

		// Declare a slice of all the stats
		statistics := make([]*Statistic, len(statsList))
		var totalCommits float64 = 0
		var totalAdditions float64 = 0
		var totalDeletions float64 = 0
		var numberAuthors int = 0

		// Loop through the slice, building the Statistics struct.
		for j := range statsList {
			// Type assert the complete stats json object.
			statsItem := statsList[j].(map[string]interface{})

			// Set the total value
			total := statsItem["total"].(float64)

			// Get the "weeks" json object.
			weeksList := statsItem["weeks"].([]interface{})

			// Create a slice for the weeks data.
			weeks := make([]*Week, len(weeksList))

			// Loop through the weeks json.
			for k := range weeksList {
				// Set the week items data.
				weekItem := weeksList[k].(map[string]interface{})
				weekNumberUnix := weekItem["w"].(float64)
				weekNumber := time.Unix(int64(weekNumberUnix), 0).Format(time.RFC3339)
				additions := weekItem["a"].(float64)
				deletions := weekItem["d"].(float64)
				commits := weekItem["c"].(float64)

				totalAdditions += additions
				totalDeletions += deletions
				totalCommits += commits

				// Add the values to the Week slice.
				weeks[int(k)] = &Week{string(weekNumber), float64(additions), float64(deletions), float64(commits)}
				//fmt.Println(string(weekNumber), float64(additions), float64(deletions), float64(commits))
			}

			author := statsItem["author"].(map[string]interface{})
			contributor := author["login"].(string)

			numberAuthors += 1

			statistics[j] = &Statistic{float64(total), weeks, string(contributor)}
			//fmt.Println(float64(total), string(contributor))
		}

		// Add the values to the Repository struct.
		repositories[i] = &Repository{repoName, bool(private), float64(size), statistics, totalCommits, totalAdditions, totalDeletions, numberAuthors}
		fmt.Fprintf(outputFile, "%s | %t | %d | %d | %d | %d | %d\n", string(repoName), bool(private), int(size), int(totalCommits), int(totalAdditions), int(totalDeletions), int(numberAuthors))
		outputFile.Sync()
	}
}
