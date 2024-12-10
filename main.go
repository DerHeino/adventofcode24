package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/joho/godotenv"
)

var DAY = "1"
var LF byte = 0x0A
var SPACE byte = 0x20

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

	// split into arrays
	columnMap := convertToIntMap(body)
	//_ = columnMap

	// sort arrays
	sortInAscendingOrder(columnMap["left"])
	sortInAscendingOrder(columnMap["right"])

	// calculate differences and sum them
	if len(columnMap["left"]) != len(columnMap["right"]) {
		log.Fatal("Both slices have different length!")
	}

	dif := 0
	right := columnMap["right"]
	for index, value := range columnMap["left"] {
		currentDif := value - right[index]
		dif += int(math.Abs(float64(currentDif)))
	}

	// return sum
	fmt.Printf("Result: %d\n", dif)
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

func convertToIntMap(b []byte) map[string][]int {

	left := make([]int, 0, 1000)
	right := make([]int, 0, 1000)

	buf := [5]byte{}
	bufIndex := 0

	for i := 0; i < len(b); i++ {

		for {
			if b[i] != LF && b[i] != SPACE {
				buf[bufIndex] = b[i]
				bufIndex++
			}

			if b[i] == SPACE {
				intValue, _ := strconv.Atoi(string(buf[:]))
				left = append(left, intValue)
				bufIndex = 0
				i = i + 2 // skip to spaces
				break
			}

			if b[i] == LF {
				intValue, _ := strconv.Atoi(string(buf[:]))
				right = append(right, intValue)
				bufIndex = 0
				break
			}

			i++
		}

		flushByteBuffer(&buf)
	}

	//fmt.Printf("%v\n%v", left, right)
	//fmt.Printf("%d, %d", len(left), len(right))

	columns := make(map[string][]int, 2)
	columns["left"] = left
	columns["right"] = right

	return columns
}

func flushByteBuffer(buf *[5]byte) {
	for i := 0; i < 5; i++ {
		buf[i] = 0x00
	}
}

func sortInAscendingOrder(slice []int) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}
