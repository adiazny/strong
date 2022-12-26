package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"golang.org/x/oauth2"
)

const version = "1.0.0"

const (
	redirectURL        = "http://localhost:4001/v1/redirect"
	stravaAuthorizeURL = "https://www.strava.com/oauth/authorize"
	stravaTokenURL     = "https://www.strava.com/oauth/token"
	stravaWriteScope   = "activity:write,activity:read"
	stravaReadScope    = "activity:read"
)

type config struct {
	port        int
	env         string
	oauthConfig *oauth2.Config
}

type application struct {
	config       config
	log          *log.Logger
	stravaClient *strava.Client
	strongConfig *strong.Config
}

// TODO
/*
	Google Drive:
	> Authorize and Authenticate
	> Get google activity for mydrive/Fitness/strong_app_workout_logs folder
	> Filter latest create/upload "strong.csv" file
	> Download latest create/upload "strong.csv" file

	Strong and Strava:
	> Process (pass) strong.csv to create strong workouts
	> Convert strong workout to strava activities
	> Post strava activities to strava apis
*/

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	var cfg config
	cfg.oauthConfig = &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  stravaAuthorizeURL,
			TokenURL: stravaTokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{stravaWriteScope},
	}

	flag.IntVar(&cfg.port, "port", 5000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.oauthConfig.ClientID, "client", os.Getenv("STRAVA_CLIENT_ID"), "Strava API Client ID")
	flag.StringVar(&cfg.oauthConfig.ClientSecret, "secret", os.Getenv("STRAVA_CLIENT_SECRET"), "Strava API Client Secret")

	flag.Parse()

	file, err := os.Open("./strong.csv")
	if err != nil {
		log.Printf("error opening file %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	records, err := strong.ReadCSV(file)
	if err != nil {
		log.Printf("error reading csv file %v\n", err)
		os.Exit(1)
	}

	workouts, err := strong.ConvertRecords(records)
	if err != nil {
		log.Printf("error converting csv to records %v\n", err)
		os.Exit(1)
	}

	completeWorkouts := strong.CombineWorkouts(workouts)

	strongConfg := &strong.Config{CompletedWorkouts: completeWorkouts}

	stravaClient := &strava.Client{Logger: log, Config: cfg.oauthConfig}

	app := &application{
		config:       cfg,
		log:          log,
		stravaClient: stravaClient,
		strongConfig: strongConfg,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("starting %s server on %s", cfg.env, srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	url := cfg.oauthConfig.AuthCodeURL("state")
	log.Println(url)

	// Blocking main.
	if err := <-serverErrors; err != nil {
		log.Fatalf("error with http server %v", err)
	}

}
