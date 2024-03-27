package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var filePath = flag.String("file", "", "the input file for the measurements.")

type Data struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int
}

func (data *Data) CalculateMean() float64 {
	ratio := math.Pow(10, float64(1))
	mean := data.Sum / float64(data.Count)
	return math.Round(mean*ratio) / ratio
}

func main() {
	flag.Parse()
	if *filePath == "" {
		log.Fatal("'-file' is required")
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	stationData := make(map[string]Data, 0)
	stationName := make([]string, 0, 50)
	lineCount := 0

	for line := range ReadFileGoRoutineV1(*filePath) {
		lineCount += 1
		lineData := strings.Split(line, ";")
		data := stationData[lineData[0]]
		measurement, err := strconv.ParseFloat(lineData[1], 64)
		if err != nil {
			log.Fatal(err)
		}

		if data.Min > measurement {
			data.Min = measurement
		}
		if data.Max < measurement {
			data.Max = measurement
		}
		data.Sum += measurement
		data.Count += 1
		stationData[lineData[0]] = data

		if lineCount%50_000_000 == 0 {
			log.Print("Parsed ", lineCount, " lines")
		}
	}

	// Put every key into a slit and sort it to make it alphabetical order
	for station := range stationData {
		stationName = append(stationName, station)
	}
	sort.Strings(stationName)

	// Put the results to stdout
	// <station name>=<min>/<mean>/<max>
	fmt.Print("{")
	for i, station := range stationName {
		data := stationData[station]
		if i == 0 {
			fmt.Printf("%s=%.1f/%.1f/%.1f", station, data.Min, data.CalculateMean(), data.Max)
		} else {
			fmt.Printf(", %s=%.1f/%.1f/%.1f", station, data.Min, data.CalculateMean(), data.Max)
		}
	}
	fmt.Println("}")

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

// Initial read implementation which is memory efficient but slow
func ReadFileGoRoutineV1(filePath string) <-chan (string) {
	line := make(chan string)
	go func() {
		defer close(line)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

	}()
	return line
}
