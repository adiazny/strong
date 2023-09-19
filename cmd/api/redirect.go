package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"golang.org/x/oauth2"
)

// redirectHandler will handle posting Strong completed workouts to Strava activities api
// if no errors then shutdown server

func parseRequest(req *http.Request) (string, error) {
	// 1) parse valid return oauth code
	urlValues := req.URL.Query()

	values, ok := urlValues["code"]
	if !ok {
		// handle query param code not in request
		return "", errors.New("something went wrong")
	}

	if len(values) > 1 {
		// handle values having more than one value
		return "", errors.New("something went wrong")
	}

	if values[0] == "" {
		// handle value being a empty string
		return "", errors.New("something went wrong")
	}

	return values[0], nil
}

func (app *application) uploadNewWorkouts(ctx context.Context, token *oauth2.Token) error {
	// 3) Get latest strava athlete activity
	stravaActivities, err := app.stravaClient.GetActivities(ctx, token)
	if err != nil {
		return err
	}

	stravaDateTimeMap := make(map[string]struct{})
	filteredStrongWorkouts := make([]strong.Workout, 0)

	// add strava dateTime to map
	for _, activity := range stravaActivities {
		stravaDateTimeMap[activity.StartDateLocal] = struct{}{}
	}

	//
	for _, strong := range app.strongConfig.CompletedWorkouts {
		if _, found := stravaDateTimeMap[strong.Date]; !found {
			// add to new map
			filteredStrongWorkouts = append(filteredStrongWorkouts, strong)
		}
	}

	// 5) convert completed workouts to strava activity
	activities := make([]strava.Actvitiy, 0)

	for _, workout := range filteredStrongWorkouts {
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
	for _, activity := range activities {
		err := app.stravaClient.PostActivity(ctx, token, activity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	// validate request:
	code, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// 2) exchange for oauth.Token
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