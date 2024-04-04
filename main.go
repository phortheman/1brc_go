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
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var filePath = flag.String("file", "", "the input file for the measurements.")

type Data map[string]*StationData

type StationData struct {
	Min   int
	Max   int
	Sum   int
	Count int
}

func (data *StationData) CalculateMean() float64 {
	ratio := math.Pow(10, float64(1))
	mean := (float64(data.Sum) / 10) / float64(data.Count)
	return math.Round(mean*ratio) / ratio
}

// Splits a line into station, measurement
func SplitLine(line string) (string, int) {
	station := ""
	isNegative := false
	tail := len(line) - 1
	measurement := 0
	decimalIndex := 0
	decimal := [3]int{1, 10, 100}
	for {
		// If it is a minus sign then the next one is a semicolon
		if line[tail] == '-' {
			station = line[:tail-1]
			isNegative = true
			break
		}
		if line[tail] == '.' {
			tail--
			continue
		}
		if line[tail] == ';' {
			station = line[:tail]
			break
		}
		digit := (int(line[tail]) - '0')
		measurement = decimal[decimalIndex]*digit + measurement
		decimalIndex++
		tail--
	}

	if isNegative {
		measurement = -1 * measurement
	}
	return station, measurement
}

/*
Put the results to stdout

<station name>=<min>/<mean>/<max>
*/
func PrintResults(data Data) {
	names := SortStations(data)
	fmt.Print("{")
	for i, name := range names {
		station := data[name]
		if i != 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s=%.1f/%.1f/%.1f", name, float64(station.Min)/10, station.CalculateMean(), float64(station.Max)/10)
	}
	fmt.Println("}")
}

// Put every key into a slice sort it to make it alphabetical order
func SortStations(data Data) []string {
	names := make([]string, 0, len(data))
	for station := range data {
		names = append(names, station)
	}
	sort.Strings(names)
	return names
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

	data := ReadDataV1(*filePath)

	PrintResults(data)

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
func ReadDataV1(filePath string) Data {
	data := make(Data)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	lineCount := 0
	for {
		readLine, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineCount += 1
		station, measurement := SplitLine(string(readLine[:]))
		stationData, ok := data[station]
		if !ok {
			data[station] = &StationData{
				Min:   measurement,
				Max:   measurement,
				Sum:   measurement,
				Count: 1,
			}
		} else {
			if stationData.Min > measurement {
				stationData.Min = measurement
			}
			if stationData.Max < measurement {
				stationData.Max = measurement
			}

			stationData.Sum += measurement
			stationData.Count += 1
		}

		if lineCount%50_000_000 == 0 {
			log.Print("Parsed ", lineCount, " lines")
		}
	}
	return data
}
