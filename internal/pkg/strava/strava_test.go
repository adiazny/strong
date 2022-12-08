package strava_test

import (
	"fmt"
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
					Weight: 44.99999997795377,
					Reps:   5},
				},
			},
			{
				Name: "Squat (Barbell)",
				Sets: []strong.Set{{
					Id:     2,
					Weight: 74.99999999999999,
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

	stravaClient := strava.Client{
		Config: &oauth2.Config{},
	}

	fmt.Println(stravaClient.MapStrongWorkout(workout))

}
