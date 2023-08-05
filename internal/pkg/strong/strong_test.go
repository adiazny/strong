package strong_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/adiazny/strong/internal/pkg/strong"
	"github.com/stretchr/testify/assert"
)

func TestReadCSV(t *testing.T) {
	t.Parallel()

	want := [][]string{{"id", "name", "age"}, {"1", "Alan", "38"}, {"2", "Carla", "34"}}

	// input = something that implements io.Reader interface
	csvAsBytes := []byte("id,name,age\n1,Alan,38\n2,Carla,34")
	buf := bytes.NewBuffer(csvAsBytes)

	// output = slice of raw records of one or more fields per record
	got, err := strong.ReadCSV(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.ElementsMatch(t, want, got)

}

func TestConvertRecords(t *testing.T) {
	t.Parallel()

	type args struct {
		records [][]string
	}
	tests := []struct {
		name    string
		args    args
		want    []strong.Workout
		wantErr bool
	}{
		{
			name: "success workout duration under one hour",
			args: args{
				records: [][]string{
					{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds", "Notes", "Workout Notes", "RPE"},
					{"2022-11-14 07:15:24", "JCDFIT Beginner A", "30m", "Squat (Barbell)", "2", "74.99999999999999", "5", "0", "0", "", "", ""},
				},
			},
			want: []strong.Workout{{
				Name:     "JCDFIT Beginner A",
				Date:     "2022-11-14 07:15:24",
				Duration: 1800000000000,
				Exercises: []strong.Exercise{{
					Name: "Squat (Barbell)",
					Sets: []strong.Set{{
						ID:     2,
						Weight: 74.99999999999999,
						Reps:   5,
					}},
				}},
			}},
			wantErr: false,
		},
		{
			name: "success workout duration at one hour",
			args: args{
				records: [][]string{
					{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds", "Notes", "Workout Notes", "RPE"},
					{"2022-11-14 07:15:24", "JCDFIT Beginner A", "1h", "Squat (Barbell)", "2", "74.99999999999999", "5", "0", "0", "", "", ""},
				},
			},
			want: []strong.Workout{{
				Name:     "JCDFIT Beginner A",
				Date:     "2022-11-14 07:15:24",
				Duration: 3600000000000,
				Exercises: []strong.Exercise{{
					Name: "Squat (Barbell)",
					Sets: []strong.Set{{
						ID:     2,
						Weight: 74.99999999999999,
						Reps:   5,
					}},
				}},
			}},
			wantErr: false,
		},
		{
			name: "success workout duration over one hour",
			args: args{
				records: [][]string{
					{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds", "Notes", "Workout Notes", "RPE"},
					{"2022-11-14 07:15:24", "JCDFIT Beginner A", "1h 30m", "Squat (Barbell)", "2", "74.99999999999999", "5", "0", "0", "", "", ""},
				},
			},
			want: []strong.Workout{{
				Name:     "JCDFIT Beginner A",
				Date:     "2022-11-14 07:15:24",
				Duration: 5400000000000,
				Exercises: []strong.Exercise{{
					Name: "Squat (Barbell)",
					Sets: []strong.Set{{
						ID:     2,
						Weight: 74.99999999999999,
						Reps:   5,
					}},
				}},
			}},
			wantErr: false,
		},
		{
			name: "success workout duration blank",
			args: args{
				records: [][]string{
					{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds", "Notes", "Workout Notes", "RPE"},
					{"2022-11-14 07:15:24", "JCDFIT Beginner A", "", "Squat (Barbell)", "2", "74.99999999999999", "5", "0", "0", "", "", ""},
				},
			},
			want: []strong.Workout{{
				Name:     "JCDFIT Beginner A",
				Date:     "2022-11-14 07:15:24",
				Duration: 0,
				Exercises: []strong.Exercise{{
					Name: "Squat (Barbell)",
					Sets: []strong.Set{{
						ID:     2,
						Weight: 74.99999999999999,
						Reps:   5,
					}},
				}},
			}},
			wantErr: false,
		},
		{
			name: "error invalid workout duration",
			args: args{
				records: [][]string{
					{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds", "Notes", "Workout Notes", "RPE"},
					{"2022-11-14 07:15:24", "JCDFIT Beginner A", "invalid workout duration", "Squat (Barbell)", "2", "74.99999999999999", "5", "0", "0", "", "", ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := strong.ConvertRecords(tt.args.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombineWorkouts(t *testing.T) {
	t.Parallel()

	type args struct {
		workouts []strong.Workout
	}

	tests := []struct {
		name string
		args args
		want []strong.Workout
	}{
		{
			name: "success with one workout day",
			args: args{
				workouts: []strong.Workout{
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     1,
									Weight: 45,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     2,
									Weight: 75,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     3,
									Weight: 95,
									Reps:   3},
								},
							},
						},
					},
				},
			},
			want: []strong.Workout{
				{
					Name:     "Day A",
					Date:     "2022-11-14 07:15:24",
					Duration: 1800000000000,
					Exercises: []strong.Exercise{
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     1,
								Weight: 45,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     2,
								Weight: 75,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     3,
								Weight: 95,
								Reps:   3},
							},
						},
					},
				},
			},
		},
		{
			name: "success with two workout days",
			args: args{
				workouts: []strong.Workout{
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     1,
									Weight: 45,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     2,
									Weight: 75,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     3,
									Weight: 95,
									Reps:   3},
								},
							},
						},
					},
					{
						Name:     "Day B",
						Date:     "2022-11-16 06:54:38",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Deadlift (Barbell)",
								Sets: []strong.Set{{
									ID:     1,
									Weight: 225,
									Reps:   8},
								},
							},
						},
					},
				},
			},
			want: []strong.Workout{
				{
					Name:     "Day A",
					Date:     "2022-11-14 07:15:24",
					Duration: 1800000000000,
					Exercises: []strong.Exercise{
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     1,
								Weight: 45,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     2,
								Weight: 75,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     3,
								Weight: 95,
								Reps:   3},
							},
						},
					},
				},
				{
					Name:     "Day B",
					Date:     "2022-11-16 06:54:38",
					Duration: 1800000000000,
					Exercises: []strong.Exercise{
						{
							Name: "Deadlift (Barbell)",
							Sets: []strong.Set{{
								ID:     1,
								Weight: 225,
								Reps:   8},
							},
						},
					},
				},
			},
		},
		{
			name: "success with two workouts same day different times",
			args: args{
				workouts: []strong.Workout{
					{
						Name:     "Day A morning",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     1,
									Weight: 45,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A morning",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     2,
									Weight: 75,
									Reps:   5},
								},
							},
						},
					},
					{
						Name:     "Day A morning",
						Date:     "2022-11-14 07:15:24",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Squat (Barbell)",
								Sets: []strong.Set{{
									ID:     3,
									Weight: 95,
									Reps:   3},
								},
							},
						},
					},
					{
						Name:     "Day A afternoon",
						Date:     "2022-11-14 15:30:00",
						Duration: 1800000000000,
						Exercises: []strong.Exercise{
							{
								Name: "Deadlift (Barbell)",
								Sets: []strong.Set{{
									ID:     1,
									Weight: 225,
									Reps:   8},
								},
							},
						},
					},
				},
			},
			want: []strong.Workout{
				{
					Name:     "Day A morning",
					Date:     "2022-11-14 07:15:24",
					Duration: 1800000000000,
					Exercises: []strong.Exercise{
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     1,
								Weight: 45,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     2,
								Weight: 75,
								Reps:   5},
							},
						},
						{
							Name: "Squat (Barbell)",
							Sets: []strong.Set{{
								ID:     3,
								Weight: 95,
								Reps:   3},
							},
						},
					},
				},
				{
					Name:     "Day A afternoon",
					Date:     "2022-11-14 15:30:00",
					Duration: 1800000000000,
					Exercises: []strong.Exercise{
						{
							Name: "Deadlift (Barbell)",
							Sets: []strong.Set{{
								ID:     1,
								Weight: 225,
								Reps:   8},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := strong.CombineWorkouts(tt.args.workouts)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestGetLatestWorkout(t *testing.T) {
	t.Parallel()

	type args struct {
		completedWorkouts []strong.Workout
	}
	tests := []struct {
		name string
		args args
		want strong.Workout
	}{
		{
			name: "Two workouts",
			args: args{
				completedWorkouts: []strong.Workout{
					{
						Name:     "Workout A",
						Date:     "2022-11-16 06:54:38",
						Duration: 60,
						Exercises: []strong.Exercise{{
							Name: "Pushup",
							Sets: []strong.Set{},
						}},
					},
					{
						Name:     "Workout B",
						Date:     "2022-11-17 06:54:38",
						Duration: 60,
						Exercises: []strong.Exercise{{
							Name: "Pullup",
							Sets: []strong.Set{},
						}},
					},
				}},
			want: strong.Workout{
				Name:     "Workout B",
				Date:     "2022-11-17 06:54:38",
				Duration: 60,
				Exercises: []strong.Exercise{{
					Name: "Pullup",
					Sets: []strong.Set{},
				}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := strong.GetLatestWorkout(tt.args.completedWorkouts)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatestWorkout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkout_Description(t *testing.T) {
	type fields struct {
		Name      string
		Date      string
		Duration  time.Duration
		Exercises []strong.Exercise
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test 1",
			fields: fields{
				Name:     "Day A",
				Date:     "Jan 16 2022",
				Duration: 3600,
				Exercises: []strong.Exercise{
					{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     1,
							Weight: 45,
							Reps:   5},
						},
					},
					{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     2,
							Weight: 75,
							Reps:   5},
						},
					},
					{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     3,
							Weight: 95,
							Reps:   3},
						},
					},
					{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     1,
							Weight: 45,
							Reps:   5},
						},
					},
					{
						Name: "Squat (Barbell)",
						Sets: []strong.Set{{
							ID:     2,
							Weight: 75,
							Reps:   5},
						},
					},
					{
						Name: "Deadlift (Barbell)",
						Sets: []strong.Set{{
							ID:     3,
							Weight: 200,
							Reps:   3},
						},
					},
					{
						Name: "Deadlift (Barbell)",
						Sets: []strong.Set{{
							ID:     3,
							Weight: 300,
							Reps:   3},
						},
					},
				},
			},
			want: `
Squat (Barbell)
Set 1: 45.0# x 5
Set 2: 75.0# x 5
Set 3: 95.0# x 3
Set 1: 45.0# x 5
Set 2: 75.0# x 5

Deadlift (Barbell)
Set 3: 200.0# x 3
Set 3: 300.0# x 3
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workout := &strong.Workout{
				Name:      tt.fields.Name,
				Date:      tt.fields.Date,
				Duration:  tt.fields.Duration,
				Exercises: tt.fields.Exercises,
			}
			if got := workout.Description(); got != tt.want {
				t.Errorf("Workout.Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDateTime(t *testing.T) {

	tests := []struct {
		name     string
		dateTime string
		want     string
		wantErr  bool
	}{
		{
			name:     "success",
			dateTime: "2022-11-13 10:48:33",
			want:     "2022-11-13T10:48:33Z",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strong.FormatDateTime(tt.dateTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println("got", got)
			if got != tt.want {
				t.Errorf("FormatDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
