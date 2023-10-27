package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/adiazny/strong/internal/pkg/strong"
	"golang.org/x/oauth2"
)

const (
	stravaBaseURL  = "https://www.strava.com/api/v3"
	activitiesPath = "activities"
	athletePath    = "athlete"
	weightTraining = "WeightTraining"
)

type Provider struct {
	*log.Logger
	*oauth2.Config
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

func (provider *Provider) GetActivities(ctx context.Context, token *oauth2.Token) ([]Actvitiy, error) {
	resp, err := provider.Client(ctx, token).Get(fmt.Sprintf("%s/%s/%s?per_page=200", stravaBaseURL, athletePath, activitiesPath))
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

		provider.Logger.Printf("%s\n", respBody)

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

func (provider *Provider) PostActivity(ctx context.Context, token *oauth2.Token, activity Actvitiy) error {
	activityData, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("error marshling activity: %w", err)
	}

	bodyReader := bytes.NewReader(activityData)

	resp, err := provider.Client(ctx, token).Post(fmt.Sprintf("%s/%s", stravaBaseURL, activitiesPath), "application/json", bodyReader)
	if err != nil {
		return fmt.Errorf("error performing http post request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error response status code is %d", resp.StatusCode)
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
