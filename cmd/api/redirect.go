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

func filterNewWorkouts(activities []strava.Actvitiy, workouts []strong.Workout) []strong.Workout {
	stravaDateTimeMap := make(map[string]struct{})
	newStrongWorkouts := make([]strong.Workout, 0)

	for _, activity := range activities {
		stravaDateTimeMap[activity.StartDateLocal] = struct{}{}
	}

	for _, strong := range workouts {
		if _, found := stravaDateTimeMap[strong.Date]; !found {
			newStrongWorkouts = append(newStrongWorkouts, strong)
		}
	}

	return newStrongWorkouts
}

func convertToStrava(workouts []strong.Workout) []strava.Actvitiy {
	newActivities := make([]strava.Actvitiy, 0)

	for _, workout := range workouts {
		activity := strava.MapStrongWorkout(workout)

		newActivities = append(newActivities, activity)
	}

	return newActivities
}

func (app *application) uploadNewWorkouts(ctx context.Context, token *oauth2.Token) error {
	stravaActivities, err := app.stravaClient.GetActivities(ctx, token)
	if err != nil {
		return err
	}

	newStrongWorkouts := filterNewWorkouts(stravaActivities, app.strongConfig.CompletedWorkouts)

	newActivities := convertToStrava(newStrongWorkouts)

	if len(newActivities) == 0 {
		return errors.New("no activities to post")
	}

	for _, activity := range newActivities {
		err := app.stravaClient.PostActivity(ctx, token, activity)
		if err != nil {
			return err
		}
	}

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
