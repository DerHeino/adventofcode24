package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var DAY = "25"

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

	parse(&body)
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

//////// Part 1 ////////

func parse(body *[]byte) {
	previousNewline := false
	row := make([]rune, 5)
	index := 0

	schematic := new(Schematic)
	keylock := new(KeyLock)

	for b, value := range *body {

		if rune(value) == '\n' && !previousNewline {
			previousNewline = true
			schematic.append(row)

			row = make([]rune, 5)
			index = 0
			if b < len(*body)-1 {
				continue
			}
		}

		if rune(value) == '\n' && previousNewline || b == len(*body)-1 {
			previousNewline = false

			schematicType := schematic.determineType()
			schematic.determineSequence()

			if schematicType == KEY {
				keylock.keys = append(keylock.keys, *schematic)
			} else {
				keylock.locks = append(keylock.locks, *schematic)
			}
			schematic = new(Schematic)
			continue
		}

		if rune(value) != '\n' {
			previousNewline = false

			row[index] = rune(value)
			index++
			continue
		}
	}

	tryToUnlock(keylock)
}

type KeyLock struct {
	keys  []Schematic
	locks []Schematic
}

type Schematic struct {
	imprint  [][]rune
	sequence []int
	isType   SchematicType
}

type SchematicType int

const (
	UNKNOWN SchematicType = 0
	KEY     SchematicType = 1
	LOCK    SchematicType = 2
)

func (s *Schematic) append(row []rune) {
	if len(s.imprint) == 0 || len(s.imprint[0]) == 0 {
		s.imprint = make([][]rune, 0, 7)
	}

	s.imprint = append(s.imprint, row)
}

func (s *Schematic) determineType() SchematicType {
	keyParity := []rune{'.', '.', '.', '.', '.'}
	lockParity := []rune{'#', '#', '#', '#', '#'}
	parity := s.imprint[0]

	if compareSlice(keyParity, parity) {
		s.isType = KEY
		return KEY
	}
	if compareSlice(lockParity, parity) {
		s.isType = LOCK
		return LOCK
	}

	log.Fatal("Schematic has neither key nor lock.")
	return -1
}

func compareSlice(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (s *Schematic) determineSequence() {
	// Not necessary
	if s.isType == UNKNOWN {
		fmt.Print("Unknown Schematic. Cannot determine Sequence")
		return
	}

	sequence := make([]int, 5)

	for _, row := range s.imprint {
		for i := 0; i < 5; i++ {
			if row[i] == '#' {
				sequence[i] += 1
			}
		}
	}

	for i := 0; i < len(sequence); i++ {
		sequence[i] = sequence[i] - 1
	}

	s.sequence = sequence
}

func tryToUnlock(keylock *KeyLock) {
	pairs := 0

	for _, key := range keylock.keys {
		for _, lock := range keylock.locks {
			pairs = pairs + isPair(key, lock)
		}
	}

	fmt.Println("Unique pairs:", pairs)
}

func isPair(key Schematic, lock Schematic) int {
	for i := 0; i < len(key.sequence); i++ {
		if key.sequence[i]+lock.sequence[i] >= 6 {
			return 0
		}
	}

	return 1
}
