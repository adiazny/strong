package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
)

// redirectHandler will handle posting Strong completed workouts to Strava activities api
// if no errors then shutdown server

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	// validate request:
	// 1) parse valid return oauth code
	urlValues := r.URL.Query()

	values, ok := urlValues["code"]
	if !ok {
		// handle query param code not in request
		http.Error(w, "", http.StatusInternalServerError)
	}

	if len(values) > 1 {
		// handle values having more than one value
		http.Error(w, "", http.StatusInternalServerError)
	}

	if values[0] == "" {
		// handle value being a empty string
		http.Error(w, "", http.StatusInternalServerError)
	}

	// 2) exchange for oauth.Token
	token, err := app.config.oauthConfig.Exchange(r.Context(), values[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 3) Get latest strava athlete activity
	stravaActivities, err := app.stravaClient.GetActivities(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//what format is the strong date?
	for _, activity := range app.strongConfig.CompletedWorkouts {
		fmt.Println("STRONG_DATE_TIME", activity.Date)
	}

	// 4) filter strong completed workouts based off latest strava activity
	filteredStrongWorkouts := strong.FilterWorkouts(app.strongConfig.CompletedWorkouts, func(workout strong.Workout) bool {
		startTime, err := time.Parse("2006-01-02 15:04:05", workout.Date)
		if err != nil {
			return false
		}

		result := false

		for _, activity := range stravaActivities {
			result = !strings.Contains(activity.StartDateLocal, startTime.String())
		}

		return result
	})

	// 5) convert completed workouts to strava activity
	activities := make([]strava.Actvitiy, 0)

	for _, workout := range filteredStrongWorkouts {
		activity, err := app.stravaClient.MapStrongWorkout(workout)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		activities = append(activities, activity)
	}

	if len(activities) == 0 {
		fmt.Fprint(w, "No workout to post to strava")

		return
	}

	// 6) post to strava api/activity
	// for _, activity := range activities {
	// 	err := app.stravaClient.PostActivity(r.Context(), token, activity)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	}
	// }

	// 7) optionally shutdown server on successful posts to strava api

	fmt.Fprint(w, "redirect successful")
}
