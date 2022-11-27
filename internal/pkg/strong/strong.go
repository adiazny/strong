package strong

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Workout struct {
	Name      string
	Date      string
	Duration  time.Duration
	Exercises []Exercise
}

type Exercise struct {
	Name string
	Sets []Set
}

type Set struct {
	Id           int
	Weight       float64
	Reps         int
	Distance     float64
	Duration     time.Duration
	Notes        string
	WorkoutNotes string
	RPE          float64
}

func ReadCSV(input io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(input)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading csv file %w", err)
	}

	return records, nil
}

func ConvertRecords(records [][]string) ([]Workout, error) {
	var workouts []Workout

	for i, record := range records {
		if i == 0 {
			continue
		}

		// dateTime, err := parseDate(record[0])
		// if err != nil {
		// 	return nil, err
		// }

		workoutDuration, err := parseWorkoutDuration(record[2])
		if err != nil {
			return nil, err
		}

		setId, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("error converting string to int for record index 4 %w", err)
		}

		weight, err := parseFloat(record[5])
		if err != nil {
			return nil, err
		}

		reps, err := strconv.Atoi(record[6])
		if err != nil {
			return nil, fmt.Errorf("error converting string to int for record index 6 %w", err)
		}

		distance, err := parseFloat(record[7])
		if err != nil {
			return nil, err
		}

		setDuration, err := parseSetDuration(record[8])
		if err != nil {
			return nil, err
		}

		rpe, err := parseFloat(record[11])
		if err != nil {
			return nil, err
		}

		workout := Workout{
			Name:     record[1],
			Date:     record[0],
			Duration: workoutDuration,
			Exercises: []Exercise{{
				Name: record[3],
				Sets: []Set{{
					Id:           setId,
					Weight:       weight,
					Reps:         reps,
					Distance:     distance,
					Duration:     setDuration,
					Notes:        record[9],
					WorkoutNotes: record[10],
					RPE:          rpe,
				}},
			}},
		}

		workouts = append(workouts, workout)
	}

	return workouts, nil
}

func CombineWorkouts(workouts []Workout) []Workout {
	if workouts == nil {
		return nil
	}

	dateWorkoutMap := make(map[string][]Workout)
	allWorkouts := make([]Workout, 0)

	for _, workout := range workouts {
		if _, ok := dateWorkoutMap[workout.Date]; !ok {
			dateWorkoutMap[workout.Date] = append(allWorkouts, workout)

			continue
		}

		if _, ok := dateWorkoutMap[workout.Date]; ok {

			dateWorkoutMap[workout.Date] = append(allWorkouts, workout)
		}
	}

	finalWorkouts := make([]Workout, 0)

	exercises := make([]Exercise, 0)

	for date, workouts := range dateWorkoutMap {
		for _, workout := range workouts {
			filteredExercises := filterExercises(workout.Exercises, func(exercise Exercise) bool {
				return workout.Date == date
			})

			exercises = append(exercises, filteredExercises...)
		}

		workout := Workout{
			Name:      workouts[0].Name,
			Date:      date,
			Duration:  workouts[0].Duration,
			Exercises: exercises,
		}

		finalWorkouts = append(finalWorkouts, workout)

	}

	return finalWorkouts
}

func parseWorkoutDuration(duration string) (time.Duration, error) {
	split := strings.Split(duration, " ")

	if split[0] == "" {
		return 0, nil
	}

	switch len(split) {
	case 1:
		min, err := strconv.Atoi(strings.TrimSuffix(split[0], "m"))
		if err != nil {
			return 0, fmt.Errorf("error converting string to int %w", err)
		}

		return time.Minute * time.Duration(min), nil
	case 2:
		hours, err := strconv.Atoi(strings.TrimSuffix(split[0], "h"))
		if err != nil {
			return 0, fmt.Errorf("error converting string to int %w", err)
		}

		minutes, err := strconv.Atoi(strings.TrimSuffix(split[1], "m"))
		if err != nil {
			return 0, fmt.Errorf("error converting string to int %w", err)
		}

		totalMinutes := (hours * 60) + minutes

		return time.Minute * time.Duration(totalMinutes), nil
	default:
		return 0, fmt.Errorf("error converting string to int for %q", duration)
	}
}

func parseSetDuration(duration string) (time.Duration, error) {
	setDurationInSeconds, err := strconv.ParseInt(duration, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("error parsing float from string for record index 8 %w", err)
	}

	return time.Duration(setDurationInSeconds) * time.Second, nil
}

func parseFloat(input string) (float64, error) {
	if input == "" {
		return 0, nil
	}

	float, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing float from string for %w", err)
	}

	return float, nil
}

func filterExercises(exercises []Exercise, matchFunc func(exercise Exercise) bool) []Exercise {
	filteredExercises := make([]Exercise, 0)

	for _, exercise := range exercises {
		match := matchFunc(exercise)

		if match {
			filteredExercises = append(filteredExercises, exercise)
		}

	}

	return filteredExercises
}
