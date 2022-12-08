package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adiazny/strong/internal/pkg/strong"
	"golang.org/x/oauth2"
)

// Model strava actvitiy
// Parse stong workout to strava activity
// Send POST request to strava api /activities

// Activity Types:
// Weight Training, Workout, Ride, Run

const weightTraining = "WeightTraining"

type Client struct {
	*log.Logger
	*oauth2.Config
}

type Actvitiy struct {
	Name           string  `json:"name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
	Elapsed_time   int     `json:"elapsed_time"`
	Description    string  `json:"description"`
	Distance       float64 `json:"distance"`
	Trainer        int     `json:"trainer"`
	Commute        int     `json:"commute"`
}

func (client *Client) MapStrongWorkout(workout strong.Workout) (Actvitiy, error) {
	time, err := time.Parse("2006-01-02 15:04:05", workout.Date)
	if err != nil {
		return Actvitiy{}, fmt.Errorf("error parsing time %w", err)
	}

	return Actvitiy{
		Name:           workout.Name,
		SportType:      weightTraining,
		StartDateLocal: time.Format("2006-01-02T15:04:05Z"),
		Elapsed_time:   int(workout.Duration.Minutes()),
		Description:    workout.String(),
	}, nil
}

func (client *Client) PostActivity(ctx context.Context, token *oauth2.Token, activity Actvitiy) error {
	activityData, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("error marshling activity: %w", err)
	}

	bodyReader := bytes.NewReader(activityData)

	resp, err := client.Client(ctx, token).Post("https://www.strava.com/api/v3/activities", "application/json", bodyReader)
	if err != nil {
		return fmt.Errorf("error performing http post request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error response status code is %d", resp.StatusCode)
	}

	return nil
}
