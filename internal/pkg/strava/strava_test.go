package strava_test

import (
	"context"
	"log"
	"reflect"
	"testing"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"golang.org/x/oauth2"
)

func TestMapStrongWorkout(t *testing.T) {
	workout := strong.Workout{
		Name:     "Day A",
		Date:     "2022-11-14 07:15:24",
		Duration: 1800000000000,
		Exercises: []strong.Exercise{
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					Id:     1,
					Weight: 45.0,
					Reps:   5},
				},
			},
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					Id:     2,
					Weight: 75.0,
					Reps:   5},
				},
			},
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					Id:     3,
					Weight: 95.0,
					Reps:   3},
				},
			},
		},
	}

	want := strava.Actvitiy{
		Name:           "Day A",
		SportType:      "WeightTraining",
		StartDateLocal: "2022-11-14T07:15:24Z",
		Elapsed_time:   1800,
		Description: `
Day A
2022-11-14 07:15:24

Squat (Barbell) Set 1: 45# x 5
Squat (Barbell) Set 2: 75# x 5
Squat (Barbell) Set 3: 95# x 3

`,
	}

	stravaClient := strava.Client{
		Config: &oauth2.Config{},
	}

	got, err := stravaClient.MapStrongWorkout(workout)
	if err != nil {
		t.Errorf("MapStrongWorkout() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("MapStrongWorkout() = %v, want %v", got, want)
	}
}

func TestClient_PostActivity(t *testing.T) {
	type fields struct {
		Logger *log.Logger
		Config *oauth2.Config
	}
	type args struct {
		ctx      context.Context
		token    *oauth2.Token
		activity strava.Actvitiy
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &strava.Client{
				Logger: tt.fields.Logger,
				Config: tt.fields.Config,
			}
			if err := client.PostActivity(tt.args.ctx, tt.args.token, tt.args.activity); (err != nil) != tt.wantErr {
				t.Errorf("Client.PostActivity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetActivities(t *testing.T) {
	type fields struct {
		Logger *log.Logger
		Config *oauth2.Config
	}
	type args struct {
		ctx   context.Context
		token *oauth2.Token
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []strava.Actvitiy
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &strava.Client{
				Logger: tt.fields.Logger,
				Config: tt.fields.Config,
			}
			got, err := client.GetActivities(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetActivities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetActivities() = %v, want %v", got, tt.want)
			}
		})
	}
}
