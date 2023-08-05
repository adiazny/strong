package strong

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	CompletedWorkouts []Workout
}

type Workout struct {
	Name      string
	Date      string
	Duration  time.Duration
	Exercises []Exercise
}

func (workout *Workout) Description() string {
	var stringBuilder strings.Builder

	exercises := make(map[string]struct{})

	for _, exercise := range workout.Exercises {
		if _, ok := exercises[exercise.Name]; !ok {
			exercises[exercise.Name] = struct{}{}
			fmt.Fprintf(&stringBuilder, "\n%s\n", exercise.Name)
		}

		for _, set := range exercise.Sets {
			fmt.Fprintf(&stringBuilder, "Set %d: %.1f# x %d", set.ID, set.Weight, set.Reps)
		}

		stringBuilder.WriteString("\n")
	}

	return stringBuilder.String()
}

type Exercise struct {
	Name string
	Sets []Set
}

type Set struct {
	ID           int
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

		dateTime, err := FormatDateTime(record[0])
		if err != nil {
			return nil, err
		}

		workoutDuration, err := parseWorkoutDuration(record[2])
		if err != nil {
			return nil, err
		}

		setID, err := strconv.Atoi(record[4])
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
			Date:     dateTime,
			Duration: workoutDuration,
			Exercises: []Exercise{{
				Name: record[3],
				Sets: []Set{{
					ID:           setID,
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

	dateSet := make(map[string]struct{})

	for _, workout := range workouts {
		if _, ok := dateSet[workout.Date]; !ok {
			dateSet[workout.Date] = struct{}{}
		}
	}

	finalWorkouts := make([]Workout, 0)

	for date := range dateSet {
		exercises := make([]Exercise, 0)

		filteredWorkouts := FilterWorkouts(workouts, func(workout Workout) bool {
			return workout.Date == date
		})

		for _, workout := range filteredWorkouts {
			exercises = append(exercises, workout.Exercises...)
		}

		workout := Workout{
			Name:      filteredWorkouts[0].Name,
			Date:      date,
			Duration:  filteredWorkouts[0].Duration,
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

	switch {
	case len(split) == 1 && strings.Contains(split[0], "m"):
		min, err := strconv.Atoi(strings.TrimSuffix(split[0], "m"))
		if err != nil {
			return 0, fmt.Errorf("error converting string to int %w", err)
		}

		return time.Minute * time.Duration(min), nil
	case len(split) == 1 && strings.Contains(split[0], "h"):
		hour, err := strconv.Atoi(strings.TrimSuffix(split[0], "h"))
		if err != nil {
			return 0, fmt.Errorf("error converting string to int %w", err)
		}

		return time.Minute * time.Duration(hour*60), nil
	case len(split) == 2:
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

func FilterWorkouts(workouts []Workout, matchFunc func(workout Workout) bool) []Workout {
	filteredWorkouts := make([]Workout, 0)

	for _, workout := range workouts {
		match := matchFunc(workout)

		if match {
			filteredWorkouts = append(filteredWorkouts, workout)
		}

	}

	return filteredWorkouts
}

func GetLatestWorkout(completedWorkouts []Workout) Workout {
	if len(completedWorkouts) == 0 {
		return Workout{}
	}

	sort.Slice(completedWorkouts, func(i, j int) bool {
		return completedWorkouts[i].Date > completedWorkouts[j].Date
	})

	return completedWorkouts[0]
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

func FormatDateTime(dateTime string) (string, error) {
	t, err := time.Parse("2006-01-02 15:04:05", dateTime)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02T15:04:05Z"), nil
}
