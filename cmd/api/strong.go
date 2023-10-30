package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adiazny/strong/internal/pkg/auth"
	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	localData "github.com/adiazny/strong/internal/pkg/strong/data"
)

const version = "1.1.0"

const (
	defaultAPIPort     = 5000
	defaultPath        = "./strong.csv"
	defaultRedirectURL = "http://localhost:4001/v1/redirect"
)

type config struct {
	port               int
	path               string
	stravaClientID     string
	stravaClientSecret string
	stravaRedirectURL  string
	gdriveClientID     string
	gdriveClientSecret string
	gdriveRedirectURL  string
}

type application struct {
	config       config
	log          *log.Logger
	stravaClient *strava.Provider
	workouts     []strong.Workout
}

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	var cfg config

	flag.IntVar(&cfg.port, "port", defaultAPIPort, "API server port")
	flag.StringVar(&cfg.path, "path", defaultPath, "Path to strong file")
	flag.StringVar(&cfg.stravaClientID, "strava-client", os.Getenv("STRAVA_CLIENT_ID"), "Strava API Client ID")
	flag.StringVar(&cfg.stravaClientSecret, "strava-secret", os.Getenv("STRAVA_CLIENT_SECRET"), "Strava API Client Secret")
	flag.StringVar(&cfg.stravaRedirectURL, "strava-redirect", defaultRedirectURL, "Strava Redirect URL")
	flag.StringVar(&cfg.gdriveClientID, "gdrive-client", os.Getenv("GDRIVE_CLIENT_ID"), "Google Drive API Client ID")
	flag.StringVar(&cfg.gdriveClientSecret, "gdrive-secret", os.Getenv("GDRIVE_CLIENT_SECRET"), "Google Drive API Client Secret")
	flag.StringVar(&cfg.gdriveRedirectURL, "gdrive-redirect", defaultRedirectURL, "Google Drive Redirect URL")
	flag.Parse()

	stravaAuth, err := auth.NewProvider(0, cfg.stravaClientID, cfg.stravaClientSecret, cfg.stravaRedirectURL)
	if err != nil {
		log.Printf("error creating straca auth provider %v\n", err)
		os.Exit(1)
	}

	//========================================================================
	// Download File from Google Drive

	// check if oauth tokens exist
	// create new drive httpClient
	// create new driveService

	// driveFileProvider := &gdrive.FileProvider{
	// 	Path: cfg.path,
	// }

	//========================================================================
	// Local File Implementation

	fp := &localData.FileProvider{
		Path: cfg.path,
	}

	fileBytes, err := fp.Import(context.Background())
	if err != nil {
		log.Printf("error importing file %v\n", err)
		os.Exit(1)
	}

	file := bytes.NewReader(fileBytes)

	workouts, err := strong.Process(file)
	if err != nil {
		log.Printf("error processing file %v\n", err)
		os.Exit(1)
	}

	//========================================================================
	// Strava bootstrap
	stravaClient := &strava.Provider{Logger: log, Config: stravaAuth}

	//========================================================================
	// API Server Setup
	app := &application{
		config:       cfg,
		log:          log,
		stravaClient: stravaClient,
		workouts:     workouts,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	serverErrors := make(chan error, 1)

	//========================================================================
	// OAuth Checks
	path, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error looking up user home directory %v", err)
	}

	filename := filepath.Join(path, "strava", "storage.json")

	token, err := tokenFromFile(filename)
	if err != nil {
		// Start api server if token file not found or errored during opening file
		go func() {
			log.Printf("starting api server on %s", srv.Addr)
			serverErrors <- srv.ListenAndServe()
		}()

		url := cfg.stravaOAuth.AuthCodeURL("state")
		log.Println(url)

		// Blocking main.
		if err := <-serverErrors; err != nil {
			log.Fatalf("error with http server %v", err)
		}
	}

	// Upload strava activites if valid token file found
	log.Print("uploading new workouts to strava")
	err = app.uploadNewWorkouts(context.Background(), token)
	if err != nil {
		log.Fatalf("error uploading strava activities %v", err)
	}
}
