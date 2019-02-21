package main

import (
	"strings"
	"time"
)

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

func main() {
	// Get the config json into the Configuration struct.
	configuration := setConfiguration()

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
