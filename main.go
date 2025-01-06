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

/*
 * -- Greedy approach --
 * Recursion probably
 * For each design
 *   Look for shortest pattern and remove that from design
 *     Repeat for shortened design
 *   Else go back to previous design and pick next shortest pattern
 *   If design is empty at the end success
 *   Else failure
 *
 * -- Reading --
 * Read first line into array of existing "patterns" should be sorted by length and letter
 * Read third line (second is empty) into array of wanted "designs"
 */

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
		return len(patterns[i]) > len(patterns[j])
	})

	designs := strings.Split(strings.Trim(string(designsByte), "\n"), "\n")

	//fmt.Printf("%v\n", patterns)
	verifyDesigns(designs, patterns)
}

func verifyDesigns(designs []string, patterns []string) {
	success := 0
	impossibleDesigns := make([]string, 0, 1024)

	for _, design := range designs {
		if verifyPatterns(design, patterns, &impossibleDesigns) {
			success++
		} else {
			fmt.Printf("[ FAILURE ] : \"%s\"\n", design)
		}
	}

	fmt.Printf("Possible designs: %d/%d\n", success, len(designs))
}

// verify if containsPatterns fails, continue until no patterns are left (that should be fine)
// if verifyPatterns fails return to last success (I think that one is broken)
func verifyPatterns(design string, patterns []string, blacklist *[]string) bool {
	for _, pattern := range containsPatterns(design, patterns, blacklist) {
		nextDesign := containsPattern(design, pattern)
		if nextDesign == "" {
			return true
		}
		if isSuccess := verifyPatterns(nextDesign, patterns, blacklist); isSuccess {
			return true
		}
	}

	*blacklist = append(*blacklist, design)
	return false
}

func containsPattern(design string, pattern string) string {
	return strings.TrimPrefix(design, pattern)
}

func containsPatterns(design string, patterns []string, blacklist *[]string) (contains []string) {
	if isImpossibleDesign(design, blacklist) {
		return
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(design, pattern) {
			contains = append(contains, pattern)
		}
	}

	return
}

func isImpossibleDesign(design string, blacklist *[]string) bool {
	for _, value := range *blacklist {
		if value == design {
			return true
		}
	}

	return false
}
