package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var DAY = "6"

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

// read input into double array, 0 0 top left
// create struct of the guard
// implement algorithm (u= -1 0, down= 1 0, left 0 -1, right 0 1)

type Field struct {
	field        [][]rune
	obsturctions int
	open         int
	walked       int
	yBorder      int
	xBorder      int
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
	yPos      int
	xPos      int
	direction rune // should be in Guard, but I'm lazy
}

type Guard struct {
	currentPos Position
	direction  Position
	path       []Position
}

const (
	UP    = '^'
	RIGHT = '>'
	DOWN  = 'v'
	LEFT  = '<'
	NONE  = 0
)

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
				guard.currentPos = Position{len(puzzleMap.field) - 1, len(row) - 1, 0}
				guard.direction = changeDirection(rune(value))
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

	runPuzzle(*puzzleMap, *guard)
}

func isGuard(symbol byte) bool {
	switch symbol {
	case byte('^'), byte('v'), byte('<'), byte('>'):
		return true
	default:
		return false
	}
}

func changeDirection(symbol rune) (direction Position) {
	switch symbol {
	case '^':
		direction.yPos = -1
		direction.direction = UP
	case 'v':
		direction.yPos = 1
		direction.direction = DOWN
	case '<':
		direction.xPos = -1
		direction.direction = LEFT
	case '>':
		direction.xPos = 1
		direction.direction = RIGHT
	}
	return
}

func nextDirection(currentDirection rune) Position {
	switch currentDirection {
	case UP:
		return changeDirection(RIGHT)
	case RIGHT:
		return changeDirection(DOWN)
	case DOWN:
		return changeDirection(LEFT)
	case LEFT:
		return changeDirection(UP)
	}
	log.Fatalf("nextDirection: Illegal argument %d!", currentDirection)
	return Position{}
}

func isObstacle(symbol rune) bool {
	return symbol == '#'
}

// Run following the algorithm:
// Walk until obstacle, turn 90 degrees if before one and repeat
// Stop if outside border

func runPuzzle(puzzleMap Field, guard Guard) {
	for {
		if !guard.isInArea(puzzleMap.yBorder, puzzleMap.xBorder) {
			break
		}

		if puzzleMap.isNextObstacle(guard) {
			guard.turn()
		} else {
			guard.walk()
		}
	}

	// welp
	for _, pos := range guard.path {
		if pos.isInArea(puzzleMap.yBorder, puzzleMap.xBorder) {
			if puzzleMap.field[pos.yPos][pos.xPos] != 'X' {
				puzzleMap.field[pos.yPos][pos.xPos] = 'X'
				puzzleMap.walked += 1
			}
		}
	}

	fmt.Printf("Distinct positions walked: %d\n", puzzleMap.walked)
}

func (f Field) isNextObstacle(g Guard) bool {
	yObstacle := g.currentPos.yPos + g.direction.yPos
	xObstacle := g.currentPos.xPos + g.direction.xPos

	obstaclePos := Position{yObstacle, xObstacle, 0}
	if obstaclePos.isInArea(f.yBorder, f.xBorder) {
		return isObstacle(f.field[yObstacle][xObstacle])
	} else {
		return false
	}
}

func (g *Guard) walk() {
	newPos := Position{g.currentPos.yPos + g.direction.yPos, g.currentPos.xPos + g.direction.xPos, 0}
	g.currentPos = newPos
	g.path = append(g.path, newPos)
}

func (g *Guard) turn() {
	g.direction = nextDirection(g.direction.direction)
}

func (g *Guard) isInArea(yBorder int, xBorder int) bool {
	return g.currentPos.isInArea(yBorder, xBorder)
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
