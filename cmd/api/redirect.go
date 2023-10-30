package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return fmt.Errorf("unable to create directory: %v", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return fmt.Errorf("unable to save oauth tokens: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)

	return nil
}

func (app *application) uploadNewWorkouts(ctx context.Context, token *oauth2.Token) error {
	stravaActivities, err := app.stravaClient.GetActivities(ctx, token)
	if err != nil {
		return err
	}

	newStrongWorkouts := filterNewWorkouts(stravaActivities, app.workouts)

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

	token, err := app.config.stravaOAuth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	path, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	filename := filepath.Join(path, "strava", "storage.json")

	err = saveToken(filename, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = app.uploadNewWorkouts(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}

	fmt.Fprint(w, "redirect successful")
}
