package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/adiazny/strong/internal/pkg/strong"
)

const (
	stravaBaseURL  = "https://www.strava.com/api/v3"
	activitiesPath = "activities"
	athletePath    = "athlete"
	weightTraining = "WeightTraining"
)

type Provider struct {
	log        *log.Logger
	httpClient *http.Client
}

func NewProvider(log *log.Logger, httpClient *http.Client) *Provider {
	return &Provider{log: log, httpClient: httpClient}
}

type Actvitiy struct {
	Name           string  `json:"name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
	ElapsedTime    int     `json:"elapsed_time"`
	Description    string  `json:"description,omitempty"`
	Distance       float64 `json:"distance"`
	Trainer        bool    `json:"trainer"`
	Commute        bool    `json:"commute"`
}

func (provider *Provider) GetActivities() ([]Actvitiy, error) {
	resp, err := provider.httpClient.Get(fmt.Sprintf("%s/%s/%s?per_page=200", stravaBaseURL, athletePath, activitiesPath))
	if err != nil {
		return nil, fmt.Errorf("error performing http get request: %w", err)
	}

	defer resp.Body.Close()

	var respBody []byte

	if resp.StatusCode != http.StatusOK {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body %w", err)
		}

		provider.log.Printf("%s\n", respBody)

		return nil, fmt.Errorf("error response status code is %d", resp.StatusCode)
	}

	activities := []Actvitiy{}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body %w", err)
	}

	err = json.Unmarshal(respBody, &activities)
	if err != nil {
		return nil, fmt.Errorf("error unmarshling response body %w", err)
	}

	return activities, nil
}

func (provider *Provider) PostActivity(activity Actvitiy) error {
	activityData, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("error marshling activity: %w", err)
	}

	bodyReader := bytes.NewReader(activityData)

	resp, err := provider.httpClient.Post(fmt.Sprintf("%s/%s", stravaBaseURL, activitiesPath), "application/json", bodyReader)
	if err != nil {
		return fmt.Errorf("error performing http post request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error response status code is %d", resp.StatusCode)
	}

	return nil
}

func (provider *Provider) UploadNewWorkouts(ctx context.Context, workouts []strong.Workout) error {
	stravaActivities, err := provider.GetActivities()
	if err != nil {
		return err
	}

	newStrongWorkouts := filterNewWorkouts(stravaActivities, workouts)

	newActivities := convertToStrava(newStrongWorkouts)

	if len(newActivities) == 0 {
		return errors.New("no strava activities to post")
	}

	for _, activity := range newActivities {
		err := provider.PostActivity(activity)
		if err != nil {
			return err
		}
	}

	return nil
}

func MapStrongWorkout(workout strong.Workout) Actvitiy {
	return Actvitiy{
		Name:           workout.Name,
		SportType:      weightTraining,
		StartDateLocal: workout.Date,
		ElapsedTime:    int(workout.Duration.Seconds()),
		Description:    workout.Description(),
	}
}

func filterNewWorkouts(activities []Actvitiy, workouts []strong.Workout) []strong.Workout {
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

func convertToStrava(workouts []strong.Workout) []Actvitiy {
	newActivities := make([]Actvitiy, 0)

	for _, workout := range workouts {
		activity := MapStrongWorkout(workout)

		newActivities = append(newActivities, activity)
	}

	return newActivities
}
