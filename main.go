package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"github.com/adiazny/strong/internal/pkg/web"
	"golang.org/x/oauth2"
)

func main() {
	file, err := os.Open("./strong.csv")
	if err != nil {
		log.Printf("error opening file %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	records, err := strong.ReadCSV(file)
	if err != nil {
		log.Printf("error reading csv file %v\n", err)
		os.Exit(1)
	}

	workouts, err := strong.ConvertRecords(records)
	if err != nil {
		log.Printf("error converting csv to records %v\n", err)
		os.Exit(1)
	}

	completeWorkouts := strong.CombineWorkouts(workouts)

	fmt.Println(completeWorkouts[0].String())

	t := template.Must(template.New("workoutLog").Parse(strong.WorkoutTemplate))

	err = t.Execute(os.Stdout, completeWorkouts)
	if err != nil {
		log.Printf("error executing text template %v\n", err)
		os.Exit(1)
	}

	// get most recent workout
	//latest := strong.GetLatestWorkout(completeWorkouts)

	// create strava client
	oauthCfg := &oauth2.Config{
		ClientID:     "59175",
		ClientSecret: "48f68c6987a6331d8400cd797e86d1c644e205b9",
		RedirectURL:  "http://localhost:3090/redirect",
		Scopes: []string{
			"activity:write",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://www.strava.com/oauth/authorize",
			TokenURL:  "https://www.strava.com/oauth/token",
			AuthStyle: 0,
		},
	}

	stravaClient := strava.NewClient(oauthCfg)

	url := stravaClient.GenerateConsentURL("state")
	log.Println(url)

	// Get Latest Strava Athlete Activities
	// filter completed workouts after latest strava activity

	// post strava activities

}
