package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
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
func printRow(r []Repository, f *os.File) {
	for i := range r {
		// Any forked repos that weren't added would have empty allocation slots at the end of the slice, so ignore these in the output.
		if r[i].Name != "" {
			fmt.Fprintf(f,
				"%s | %s | %d | %d | %d | %d | %d\n",
				r[i].Name,
				r[i].Visibility,
				r[i].Size,
				r[i].TotalCommits,
				r[i].TotalAdditions,
				r[i].TotalDeletions,
				r[i].NumberAuthors)
		}
	}
}

// outputMarkdown sends the repository list to the markdown file.
func outputMarkdown(repositories []Repository) {
	// Create the output file.
	outputFile, err := os.Create("README.md")
	check(err)
	defer outputFile.Close()

	// Add the headers.
	fmt.Fprint(outputFile, getMarkdownTemplate())
	outputFile.Sync()

	// Closures to order the Repository structure.
	size := func(r1, r2 *Repository) bool {
		return r2.Size < r1.Size
	}

	// Sort the repositories by Size.
	By(size).Sort(repositories)

	// Print out each row to the README.md
	printRow(repositories, outputFile)
	outputFile.Sync()
}
