package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var DAY = "6"
var PART_TWO = true

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

	startPuzzle(body)
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

// double array, 0(y) 0(x) top left
// coordinate in a field is called "tile"
type Field struct {
	field        [][]rune
	obsturctions int
	open         int
	yBorder      int
	xBorder      int
	distinct     int
}

func (f *Field) isObstacle(yPos int, xPos int) bool {
	symbol := f.field[yPos][xPos]
	return symbol == '#' || symbol == 'O'
}

func (f *Field) isOpen(yPos int, xPos int) bool {
	return f.field[yPos][xPos] == '.'
}

func (f *Field) isNextObstacle(g Guard) bool {
	yObstacle := g.currentPos.yPos + g.step.yPos
	xObstacle := g.currentPos.xPos + g.step.xPos

	obstaclePos := Position{yObstacle, xObstacle}
	if obstaclePos.isInArea(f.yBorder, f.xBorder) {
		return f.isObstacle(yObstacle, xObstacle)
	} else {
		return false
	}
}

func (f Field) toString() (s string) {
	for _, yValue := range f.field {
		for _, xValue := range yValue {
			s = s + string(xValue) + " "
		}
		s = s + "\n"
	}
	return
}

type Position struct {
	yPos int
	xPos int
}

func (p *Position) isInArea(yBorder int, xBorder int) bool {
	if p.yPos < 0 || p.yPos >= yBorder {
		return false
	}
	if p.xPos < 0 || p.xPos >= xBorder {
		return false
	}
	return true
}

const (
	UP    = '^'
	RIGHT = '>'
	DOWN  = 'v'
	LEFT  = '<'
	NONE  = 0
)

type Guard struct {
	currentPos Position
	direction  rune
	step       Position
	path       []Position
}

func (g *Guard) isInArea(yBorder int, xBorder int) bool {
	return g.currentPos.isInArea(yBorder, xBorder)
}

func (g *Guard) changeDirection(direction rune) {
	switch g.direction = direction; g.direction {
	case '^':
		g.step.yPos = -1
		g.step.xPos = 0
	case 'v':
		g.step.yPos = 1
		g.step.xPos = 0
	case '<':
		g.step.yPos = 0
		g.step.xPos = -1
	case '>':
		g.step.yPos = 0
		g.step.xPos = 1
	default:
		log.Fatalf("chageDirection: Illegal argument '%s'", string(direction))
	}
}

func (g *Guard) turn() {
	switch g.direction {
	case UP:
		g.changeDirection(RIGHT)
	case RIGHT:
		g.changeDirection(DOWN)
	case DOWN:
		g.changeDirection(LEFT)
	case LEFT:
		g.changeDirection(UP)
	}
}

func (g *Guard) walk() {
	newPos := Position{g.currentPos.yPos + g.step.yPos, g.currentPos.xPos + g.step.xPos}
	g.currentPos = newPos
	g.path = append(g.path, newPos)
}

func (g *Guard) preview() Position {
	return Position{g.currentPos.yPos + g.step.yPos, g.currentPos.xPos + g.step.xPos}
}

// initialize Field and Guard with start Position
func startPuzzle(body []byte) {
	var puzzleMap = new(Field)
	var guard = new(Guard)

	row := make([]rune, 0, 200)
	for _, value := range body {
		if value != byte('\n') {
			row = append(row, rune(value))
			switch {
			case value == byte('#'):
				puzzleMap.obsturctions += 1
			case value == byte('.'):
				puzzleMap.open += 1
			case isGuard(value):
				guard.currentPos = Position{len(puzzleMap.field) - 1, len(row) - 1}
				guard.changeDirection(rune(value))
				guard.path = make([]Position, 200)
				guard.path = append(guard.path, guard.currentPos)
			}
		} else {
			puzzleMap.field = append(puzzleMap.field, row)
			row = make([]rune, 0, 200)
		}
	}

	puzzleMap.yBorder = len(puzzleMap.field)
	puzzleMap.xBorder = len(puzzleMap.field[0])

	//fmt.Println(puzzleMap.toString())

	if !PART_TWO {
		runPuzzle(*puzzleMap, *guard)
	} else {
		runPuzzlePart2(*puzzleMap, *guard)
	}
}

func isGuard(symbol byte) bool {
	switch symbol {
	case byte('^'), byte('v'), byte('<'), byte('>'):
		return true
	default:
		return false
	}
}

// runs algorithm:
// Guard walks until obstacle, then turns 90 degrees
// ends if Guard is outside Field and calculates distinct tiles walked inside Field
func runPuzzle(puzzleMap Field, guard Guard) bool {
	start := time.Now()
	for {
		if !guard.isInArea(puzzleMap.yBorder, puzzleMap.xBorder) {
			break
		}

		if puzzleMap.isNextObstacle(guard) {
			guard.turn()
		} else {
			guard.walk()
		}

		// this actually works lol
		if time.Until(start) < -(200 * time.Millisecond) {
			return true
		}
	}

	if !PART_TWO {
		calulateDistinctTiles(puzzleMap, guard)
		return false
	}

	return false
}

func calulateDistinctTiles(puzzleMap Field, guard Guard) {
	// welp
	for _, pos := range guard.path {
		if pos.isInArea(puzzleMap.yBorder, puzzleMap.xBorder) {
			if puzzleMap.field[pos.yPos][pos.xPos] != 'X' {
				puzzleMap.field[pos.yPos][pos.xPos] = 'X'
				puzzleMap.distinct += 1
			}
		}
	}

	fmt.Printf("Distinct positions walked: %d\n", puzzleMap.distinct)
}

//////// PART 2 ////////

func runPuzzlePart2(puzzleMap Field, guard Guard) {
	counter := 0

	for y, row := range puzzleMap.field {
		for x := range row {
			if puzzleMap.isOpen(y, x) {
				puzzleMap.field[y][x] = 'O'

				if runPuzzle(puzzleMap, guard) {
					counter += 1
				}

				puzzleMap.field[y][x] = '.'
			}
		}
	}

	fmt.Printf("Possible Obstacles: %d", counter)
}
