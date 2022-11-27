package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

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

	completeWorkouts := strong.CombineWorkouts(workouts)

	fmt.Println(completeWorkouts[0].String())

	t := template.Must(template.New("workoutLog").Parse(strong.WorkoutTemplate))

	err = t.Execute(os.Stdout, completeWorkouts)
	if err != nil {
		log.Println("executing template:", err)
	}

}
