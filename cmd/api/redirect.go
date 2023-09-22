package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"github.com/adiazny/strong/internal/pkg/web"
	"golang.org/x/oauth2"
)

func (app *application) filterNewWorkouts(activities []strava.Actvitiy) []strong.Workout {
	stravaDateTimeMap := make(map[string]struct{})
	filteredStrongWorkouts := make([]strong.Workout, 0)

	// add strava dateTime to map
	for _, activity := range activities {
		stravaDateTimeMap[activity.StartDateLocal] = struct{}{}
	}

	for _, strong := range app.strongConfig.CompletedWorkouts {
		if _, found := stravaDateTimeMap[strong.Date]; !found {
			// add to new map
			filteredStrongWorkouts = append(filteredStrongWorkouts, strong)
		}
	}

	return filteredStrongWorkouts
}

func (app *application) uploadNewWorkouts(ctx context.Context, token *oauth2.Token) error {
	// 3) Get latest strava athlete activity
	stravaActivities, err := app.stravaClient.GetActivities(ctx, token)
	if err != nil {
		return err
	}

	newStrongWorkouts := app.filterNewWorkouts(stravaActivities)

	fmt.Printf("Strava Activity %+v \n", stravaActivities[0])

	fmt.Printf("Strong Workout %+v \n", app.strongConfig.CompletedWorkouts[0])

	// 5) convert completed workouts to strava activity
	activities := make([]strava.Actvitiy, 0)

	for _, workout := range newStrongWorkouts {
		activity, err := app.stravaClient.MapStrongWorkout(workout)
		if err != nil {
			return err
		}

		activities = append(activities, activity)
	}

	if len(activities) == 0 {
		return errors.New("no workout to post")
	}

	// 6) post to strava api/activity
	// for _, activity := range activities {
	// 	err := app.stravaClient.PostActivity(ctx, token, activity)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	code, err := web.ParseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	token, err := app.config.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = app.uploadNewWorkouts(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}

	fmt.Fprint(w, "redirect successful")
}
