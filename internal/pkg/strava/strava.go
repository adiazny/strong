package strava

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// Model strava actvitiy
// Parse stong workout to strava activity
// Send POST request to strava api /activities

// Activity Types:
// Weight Training, Workout, Ride, Run

const (
	Crossfit SportType = iota
	Hike
	Ride
	Run
	Swim
	TrailRun
	VirtualRide
	VirtualRun
	Walk
	WeightTraining
	Workout
	Yoga
)

type SportType int

type Client struct {
	*oauth2.Config
}

type Actvitiy struct {
	Name           string    `json:"name"`
	SportType      SportType `json:"sport_type"`
	StartDateLocal string    `json:"start_date_local"`
	Elapsed_time   int       `json:"elapsed_time"`
	Description    string    `json:"description"`
	Distance       float64   `json:"distance"`
	Trainer        int       `json:"trainer"`
	Commute        int       `json:"commute"`
}

func NewClient(config *oauth2.Config) *Client {
	// url redirect to get user consent
	// get tokens
	// return http.client

	return &Client{config}
}

func (client *Client) GenerateConsentURL(state string) string {
	return client.AuthCodeURL(state)
}

func (client *Client) TokenExchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := client.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error exchanging token %w", err)
	}

	return token, nil
}

func (client *Client) GetClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return client.Client(ctx, token)
}
