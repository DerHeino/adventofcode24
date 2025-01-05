package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

var DAY = "8"

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

	readIntoFieldAndStartPuzzle(body)
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

//////// PART 1 implementation ////////

type Field struct {
	antennaField  [][]rune
	antinodeField [][]rune
	yBorder       int
	xBorder       int
	antennaMap    map[rune][]Location
}

func (f *Field) addAntenna(key rune, value Location) {
	if f.antennaMap[key] == nil {
		f.antennaMap[key] = make([]Location, 0, 50)
	}

	f.antennaMap[key] = append(f.antennaMap[key], value)
}

func (f *Field) addAntinode(value Location) bool {
	if f.isInField(value.y, value.x) {
		f.antinodeField[value.y][value.x] = '#'
		return true
	}
	return false
}

func (f *Field) isInField(y int, x int) bool {
	if y < 0 || y >= f.yBorder {
		return false
	}
	if x < 0 || x >= f.xBorder {
		return false
	}
	return true
}

func (f *Field) amountOfAntinodes() (amount int) {
	for y := 0; y < f.yBorder; y++ {
		for x := 0; x < f.xBorder; x++ {
			if f.antinodeField[y][x] == '#' {
				amount++
			}

			if f.antennaField[y][x] != '.' && f.antinodeField[y][x] != '#' {
				amount++
			}
		}
	}

	return
}

func (f *Field) antennaFieldToString() string {
	return f.fieldToString(f.antennaField)
}

func (f *Field) antinodeFieldToString() string {
	return f.fieldToString(f.antinodeField)
}

func (f *Field) fieldToString(field [][]rune) (output string) {
	for _, rows := range field {
		output += "[ "
		for _, value := range rows {
			output += string(value)
		}
		output += " ]\n"
	}

	return
}

type Location struct {
	y int
	x int
}

func (direction *Location) getInverse() Location {
	switch *direction {
	case TOP:
		return BOTTOM
	case BOTTOM:
		return TOP
	case LEFT:
		return RIGHT
	case RIGHT:
		return LEFT
	case TOP_LEFT:
		return BOTTOM_RIGHT
	case TOP_RIGHT:
		return BOTTOM_LEFT
	case BOTTOM_LEFT:
		return TOP_RIGHT
	case BOTTOM_RIGHT:
		return TOP_LEFT
	default:
		log.Fatalf("Illegal direction given! Direction: %v", *direction)
		return Location{}
	}
}

var (
	TOP    = Location{-1, 0}
	BOTTOM = Location{1, 0}
	LEFT   = Location{0, -1}
	RIGHT  = Location{0, 1}

	TOP_LEFT     = Location{-1, -1}
	TOP_RIGHT    = Location{-1, 1}
	BOTTOM_LEFT  = Location{1, -1}
	BOTTOM_RIGHT = Location{1, 1}

	EQUAL = Location{}
)

func readIntoFieldAndStartPuzzle(body []byte) {
	var puzzle = new(Field)
	var alphanumeric, _ = regexp.Compile("(^[0-9A-Za-z]{1}$)")

	puzzle.antennaMap = make(map[rune][]Location, 62)
	row := make([]rune, 0, 100)

	for _, value := range body {
		if value != byte('\n') {
			row = append(row, rune(value))
			if alphanumeric.Match([]byte{value}) {
				puzzle.addAntenna(rune(value), Location{len(puzzle.antennaField), len(row) - 1})
			}
		} else {
			puzzle.antennaField = append(puzzle.antennaField, row)
			puzzle.antinodeField = append(puzzle.antinodeField, make([]rune, len(row), 100))
			row = make([]rune, 0, 100)
		}
	}

	puzzle.yBorder = len(puzzle.antennaField)
	puzzle.xBorder = len(puzzle.antennaField[0])

	for yIndex := 0; yIndex < puzzle.yBorder; yIndex++ {
		for xIndex := 0; xIndex < puzzle.xBorder; xIndex++ {
			puzzle.antinodeField[yIndex][xIndex] = '.'
		}
	}

	calculateAllFrequencies(*puzzle)

	fmt.Println(puzzle.antennaFieldToString())
	fmt.Println(puzzle.antinodeFieldToString())
	fmt.Printf("Amount of antinodes in field: %d.\n", puzzle.amountOfAntinodes())
}

func calculateAllFrequencies(puzzle Field) {
	for _, antennas := range puzzle.antennaMap {
		calculateFrequencies(antennas, puzzle)
	}
}

func calculateFrequencies(antennas []Location, puzzle Field) {
	for index, antenna1 := range antennas {
		if index == len(antennas) {
			break
		}

		for i := index + 1; i < len(antennas); i++ {
			antenna2 := antennas[i]

			distance, multiplier1, multiplier2 := calculateDirection(antenna1, antenna2)

			calculateUntilOutOfField(puzzle, antenna1, distance, multiplier1)
			calculateUntilOutOfField(puzzle, antenna2, distance, multiplier2)
		}
	}
}

func calculateDirection(location1 Location, location2 Location) (distance Location, multiplier1 Location, multiplier2 Location) {
	var yDirection Location
	var xDirection Location

	yDifference := location1.y - location2.y
	xDifference := location1.x - location2.x

	if yDifference < 0 {
		yDirection = TOP
	} else if yDifference > 0 {
		yDirection = BOTTOM
	} else {
		yDirection = EQUAL
	}

	if xDifference < 0 {
		xDirection = LEFT
	} else if xDifference > 0 {
		xDirection = RIGHT
	} else {
		xDirection = EQUAL
	}

	if yDirection == EQUAL {
		multiplier1 = xDirection
		multiplier2 = multiplier1.getInverse()
	} else if xDirection == EQUAL {
		multiplier1 = yDirection
		multiplier2 = multiplier1.getInverse()
	} else if yDirection == TOP && xDirection == LEFT {
		multiplier1 = TOP_LEFT
		multiplier2 = multiplier1.getInverse()
	} else if yDirection == TOP && xDirection == RIGHT {
		multiplier1 = TOP_RIGHT
		multiplier2 = multiplier1.getInverse()
	} else if yDirection == BOTTOM && xDirection == LEFT {
		multiplier1 = BOTTOM_LEFT
		multiplier2 = multiplier1.getInverse()
	} else if yDirection == BOTTOM && xDirection == RIGHT {
		multiplier1 = BOTTOM_RIGHT
		multiplier2 = multiplier1.getInverse()
	}

	distance.y, distance.x = int(math.Abs(float64(yDifference))), int(math.Abs(float64(xDifference)))
	return
}

func calculateAntinode(antenna Location, distance Location, multiplier Location) Location {
	y := antenna.y + (distance.y * multiplier.y)
	x := antenna.x + (distance.x * multiplier.x)

	return Location{y, x}
}

//////// PART 2 implementation ////////

func calculateUntilOutOfField(puzzle Field, antenna Location, distance Location, multiplier Location) {
	currentNode := antenna

	for {
		nextNode := calculateAntinode(currentNode, distance, multiplier)
		if puzzle.addAntinode(nextNode) {
			currentNode = nextNode
		} else {
			break
		}
	}
}
