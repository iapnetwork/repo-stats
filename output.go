package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

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

// getMarkdownTemplate reads and returns the contents of the markdown template file.
func getMarkdownTemplate() string {
	// Read the template data from the file.
	data, err := ioutil.ReadFile("template.md")
	check(err)

	return string(data)
}

// printRow outputs the contents of each repository statistics to the README.
func outputTable(r []Repository) string {
	var table string
	var repos, size, authors int
	var commits, additions, deletions int64

	for _, item := range r {
		// Any forked repos that weren't added would have empty allocation slots at the end of the slice, so ignore these in the output.
		if item.Name != "" {
			table += fmt.Sprintf(
				"| %s | %s | %d | %d | %d | %d | %d |\n",
				item.Name,
				item.Visibility,
				item.Size,
				item.TotalStats.Commits,
				item.TotalStats.Additions,
				item.TotalStats.Deletions,
				item.TotalStats.Authors)

			// Calculate totals.
			repos += 1
			size += item.Size
			commits += item.TotalStats.Commits
			additions += item.TotalStats.Additions
			deletions += item.TotalStats.Deletions
			if authors < item.TotalStats.Authors {
				authors = item.TotalStats.Authors
			}
		}
	}

	// Add totals to the end of the table.
	table += fmt.Sprint("| | | | | | | |\n")
	table += fmt.Sprintf(
		"| **Totals** | **%d** | **%d** | **%d** | **%d** | **%d** | **%d** |\n",
		repos,
		size,
		commits,
		additions,
		deletions,
		authors)

	return table
}

// outputMarkdown sends the repository list to the markdown file.
func outputMarkdown(repositories []Repository) {
	// Create the output file.
	readme, err := os.Create("README.md")
	check(err)
	defer readme.Close()

	// Add the headers.
	template := fmt.Sprint(getMarkdownTemplate())

	// Closures to order the Repository structure.
	size := func(r1, r2 *Repository) bool {
		return r2.Size < r1.Size
	}

	// Sort the repositories by Size.
	By(size).Sort(repositories)

	// Print out all of the rows to the README.md
	output := strings.Replace(template, "{{ table }}", outputTable(repositories), 1)
	_, err = fmt.Fprint(readme, output)
	check(err)
	check(readme.Sync())
}
