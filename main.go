package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

var DAY = "19"

func main() {

	// consume session
	session := retrieveSession()

	// fetch file from www
	url := "https://adventofcode.com/2024/day/" + DAY + "/input"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Cookie", "session="+session)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Fatal("Input could not be fetched: " + resp.Status)
	}

	// no response body length check as it is in HTTP/2.0 and no Content-Length is present in the header

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read input stream!")
	}

	resp.Body.Close()
	defer resp.Body.Close()

	parse(body)
}

func retrieveSession() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("No .env file found!")
	}

	key, exists := os.LookupEnv("session")
	if !exists {
		log.Fatal("No session token found!")
	}

	return key
}

//////// Day 19 implementation ////////

type Patterns struct {
	templates     []string
	blacklist     []string
	whitelist     []string
	cache         map[string]int
	possibilities int
}

func (p *Patterns) appendToBlacklist(impossible string) {
	p.blacklist = append(p.blacklist, impossible)
}

func (p *Patterns) appendToWhitelist(design string) {
	p.whitelist = append(p.whitelist, design)
}

func parse(body []byte) {
	var patternsByte []byte
	var designsByte []byte

	var patternFlag = true

	for i, value := range body {

		if patternFlag && rune(value) == '\n' {
			patternsByte = body[:i]
			patternFlag = false
			continue
		} else if !patternFlag && rune(value) != '\n' {
			designsByte = body[i:]
			break
		}
	}

	patterns := strings.Split(string(patternsByte), ", ")
	sort.Slice(patterns, func(i, j int) bool {
		if len(patterns[i]) == len(patterns[j]) {
			return patterns[i] < patterns[j]
		}
		return len(patterns[i]) < len(patterns[j])
	})

	designs := strings.Split(strings.Trim(string(designsByte), "\n"), "\n")

	possibleDesigns := verifyDesigns(designs, patterns)
	calculatePossibilitiesForAll(&possibleDesigns)
}

func verifyDesigns(designs []string, templates []string) Patterns {
	patterns := Patterns{templates, make([]string, 0, 1024), make([]string, 0, 400), make(map[string]int, 4000), 0}
	success := 0

	for _, design := range designs {
		if designPossible(design, &patterns, 0) {
			patterns.appendToWhitelist(design)
			success++
		} else {
			patterns.appendToBlacklist(design)
		}
	}

	fmt.Printf("Possible designs: %d/%d\n", success, len(designs))
	return patterns
}

// verify if containsPatterns fails, continue until no patterns are left (that should be fine)
// if designPossible fails return to last success (I think that one is broken)
func designPossible(design string, patterns *Patterns, depth int) bool {
	status := false

	for _, pattern := range getPossiblePatterns(design, patterns) {
		nextDesign := strings.TrimPrefix(design, pattern)
		if nextDesign == "" {
			return true
		}
		if possible := designPossible(nextDesign, patterns, depth+1); possible {
			return true
		}
	}

	patterns.appendToBlacklist(design)
	return status
}

func getPossiblePatterns(design string, patterns *Patterns) (contains []string) {
	if isImpossibleDesign(design, patterns.blacklist) {
		return
	}

	for _, pattern := range patterns.templates {
		if strings.HasPrefix(design, pattern) {
			contains = append(contains, pattern)
		}
	}

	return
}

func isImpossibleDesign(design string, blacklist []string) bool {
	for _, value := range blacklist {
		if value == design {
			return true
		}
	}

	return false
}

//////// Part 2 implementation ////////

func calculatePossibilitiesForAll(patterns *Patterns) Patterns {

	for _, value := range patterns.whitelist {
		patterns.possibilities += calculatePossibilities(value, patterns, 0)
	}

	fmt.Printf("Combinations of possible designs: %d\n", patterns.possibilities)
	return *patterns
}

func calculatePossibilities(design string, patterns *Patterns, depth int) int {
	if _, in := patterns.cache[design]; in {
		return patterns.cache[design]
	}

	for _, pattern := range getPossiblePatterns(design, patterns) {
		if strings.HasPrefix(design, pattern) {
			if nextDesign := strings.TrimPrefix(design, pattern); nextDesign == "" {
				patterns.cache[design] = patterns.cache[design] + 1
			} else {
				patterns.cache[design] = patterns.cache[design] + calculatePossibilities(nextDesign, patterns, depth+1)
			}
		}
	}

	return patterns.cache[design]
}
