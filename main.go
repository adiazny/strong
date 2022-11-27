package main

import (
	"fmt"
	"os"

	"github.com/adiazny/strong/internal/pkg/strong"
)

func main() {

	file, err := os.Open("./strong.csv")
	if err != nil {
		os.Exit(1)
	}

	defer file.Close()

	records, err := strong.ReadCSV(file)
	if err != nil {
		os.Exit(1)
	}

	workouts, err := strong.ConvertRecords(records)
	if err != nil {
		os.Exit(1)
	}

	for _, workout := range workouts {
		fmt.Printf("Workout %+v", workout)
	}
}
