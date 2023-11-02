package strava_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
)

func TestMapStrongWorkout(t *testing.T) {
	workout := strong.Workout{
		Name:     "Day A",
		Date:     "2022-11-14T07:15:24Z",
		Duration: 1800000000000,
		Exercises: []strong.Exercise{
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					ID:     1,
					Weight: 45.0,
					Reps:   5},
				},
			},
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					ID:     2,
					Weight: 75.0,
					Reps:   5},
				},
			},
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					ID:     3,
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
		ElapsedTime:    1800,
		Description: `
Squat (Barbell)
Set 1: 45.0# x 5
Set 2: 75.0# x 5
Set 3: 95.0# x 3
`,
	}

	got := strava.MapStrongWorkout(workout)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("MapStrongWorkout() = %v, want %v", got, want)
	}
}

func TestProvider_PostActivity(t *testing.T) {
	type mockArgs struct {
		httpStatusCode int
		httpResponse   string
	}

	type args struct {
		activity strava.Actvitiy
	}

	tests := []struct {
		name     string
		mockArgs mockArgs
		args     args
		wantErr  bool
	}{
		{
			name:     "success",
			mockArgs: mockArgs{httpStatusCode: http.StatusCreated, httpResponse: ""},
			args:     args{activity: strava.Actvitiy{}},
			wantErr:  false,
		},
		{
			name:     "invalid http status code of 409",
			mockArgs: mockArgs{httpStatusCode: http.StatusConflict, httpResponse: ""},
			args:     args{activity: strava.Actvitiy{}},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := log.New(os.Stdout, "", log.Ldate|log.Ltime)
			client := &http.Client{}

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockArgs.httpStatusCode)
				fmt.Fprint(w, tt.mockArgs.httpResponse)
			}))

			defer svr.Close()

			provider := strava.NewProvider(log, svr.URL, client)

			if err := provider.PostActivity(tt.args.activity); (err != nil) != tt.wantErr {
				t.Errorf("provider.PostActivity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_GetActivities(t *testing.T) {
	type mockArgs struct {
		httpStatusCode int
		httpResponse   string
	}

	tests := []struct {
		name     string
		mockArgs mockArgs
		want     []strava.Actvitiy
		wantErr  bool
	}{
		{
			name:     "success",
			mockArgs: mockArgs{httpStatusCode: http.StatusOK, httpResponse: mockStravaActivitiesJSON()},
			want: []strava.Actvitiy{
				{
					Name:           "Happy Friday",
					SportType:      "MountainBikeRide",
					StartDateLocal: "2018-05-02T05:15:09Z",
					ElapsedTime:    4500,
					Distance:       24931.4,
				},
				{
					Name:           "Bondcliff",
					SportType:      "MountainBikeRide",
					StartDateLocal: "2018-04-30T05:35:51Z",
					ElapsedTime:    5400,
					Distance:       23676.5,
				},
			},
			wantErr: false,
		},
		{
			name:     "401 http status code",
			mockArgs: mockArgs{httpStatusCode: http.StatusUnauthorized, httpResponse: ""},
			wantErr:  true,
		},
		{
			name:     "bad json",
			mockArgs: mockArgs{httpStatusCode: http.StatusOK, httpResponse: "bad json"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := log.New(os.Stdout, "", log.Ldate|log.Ltime)
			client := &http.Client{}

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockArgs.httpStatusCode)
				fmt.Fprint(w, tt.mockArgs.httpResponse)
			}))

			defer svr.Close()

			provider := strava.NewProvider(log, svr.URL, client)

			got, err := provider.GetActivities()
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.GetActivities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.GetActivities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_UploadNewWorkouts(t *testing.T) {
	type mockArgs struct {
		httpStatusCode int
		httpResponse   string
	}
	type args struct {
		ctx      context.Context
		workouts []strong.Workout
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		args     args
		wantErr  bool
	}{
		{
			name:     "success",
			mockArgs: mockArgs{httpStatusCode: http.StatusOK, httpResponse: mockStravaActivitiesJSON()},
			args: args{
				context.Background(),
				[]strong.Workout{{
					Name:     "JCDFIT Beginner A",
					Date:     "2022-11-14T07:15:24Z",
					Duration: 5400000000000,
					Exercises: []strong.Exercise{{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     2,
							Weight: 75,
							Reps:   5,
						}},
					}},
				}},
			},
			wantErr: false,
		},
		{
			name:     "no new activities to post",
			mockArgs: mockArgs{httpStatusCode: http.StatusOK, httpResponse: mockStravaActivitiesJSON()},
			args: args{
				context.Background(),
				[]strong.Workout{{
					Name:     "JCDFIT Beginner A",
					Date:     "2018-04-30T05:35:51Z",
					Duration: 5400000000000,
					Exercises: []strong.Exercise{{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     2,
							Weight: 75,
							Reps:   5,
						}},
					}},
				}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := log.New(os.Stdout, "", log.Ldate|log.Ltime)
			client := &http.Client{}

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
				}

				if r.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
				}

				fmt.Fprint(w, tt.mockArgs.httpResponse)
			}))

			defer svr.Close()

			provider := strava.NewProvider(log, svr.URL, client)

			if err := provider.UploadNewWorkouts(tt.args.ctx, tt.args.workouts); (err != nil) != tt.wantErr {
				t.Errorf("provider.UploadNewWorkouts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockStravaActivitiesJSON() string {
	return `
	[
		{
		  "resource_state": 2,
		  "athlete": {
			"id": 134815,
			"resource_state": 1
		  },
		  "name": "Happy Friday",
		  "distance": 24931.4,
		  "moving_time": 4500,
		  "elapsed_time": 4500,
		  "type": "Ride",
		  "sport_type": "MountainBikeRide",
		  "id": 154504250376823,
		  "start_date": "2018-05-02T12:15:09Z",
		  "start_date_local": "2018-05-02T05:15:09Z",
		  "timezone": "(GMT-08:00) America/Los_Angeles"
		},
		{
		  "resource_state": 2,
		  "athlete": {
			"id": 167560,
			"resource_state": 1
		  },
		  "name": "Bondcliff",
		  "distance": 23676.5,
		  "moving_time": 5400,
		  "elapsed_time": 5400,
		  "type": "Ride",
		  "sport_type": "MountainBikeRide",
		  "id": 1234567809,
		  "start_date": "2018-04-30T12:35:51Z",
		  "start_date_local": "2018-04-30T05:35:51Z",
		  "timezone": "(GMT-08:00) America/Los_Angeles"
		}
	  ]`
}
