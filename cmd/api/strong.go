package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	localData "github.com/adiazny/strong/internal/pkg/strong/data"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/driveactivity/v2"
)

const version = "1.1.0"

const (
	defaultAPIPort     = 5000
	defaultPath        = "./strong.csv"
	defaultRedirectURL = "http://localhost:4001/v1/redirect"
	stravaAuthorizeURL = "https://www.strava.com/oauth/authorize"
	stravaTokenURL     = "https://www.strava.com/oauth/token"
	stravaScopes       = "activity:write,activity:read"
)

type config struct {
	port        int
	path        string
	stravaOAuth *oauth2.Config
	gdriveOAuth *oauth2.Config
}

type application struct {
	config       config
	log          *log.Logger
	stravaClient *strava.Provider
	workouts     []strong.Workout
}

// TODO
/*
	Google Drive:
	> Authorize and Authenticate
	>> Look into service accounts https://developers.google.com/identity/protocols/oauth2/service-account
	> Get google activity for mydrive/Fitness/strong_app_workout_logs folder
	> Filter latest create/upload "strong.csv" file
	> Download latest create/upload "strong.csv" file

*/

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}

	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	var cfg config

	cfg.stravaOAuth = &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  stravaAuthorizeURL,
			TokenURL: stravaTokenURL,
		},
		Scopes: []string{stravaScopes},
	}

	cfg.gdriveOAuth = &oauth2.Config{
		Scopes: []string{
			drive.DriveReadonlyScope,
			driveactivity.DriveActivityReadonlyScope,
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: "http://localhost",
	}

	flag.IntVar(&cfg.port, "port", defaultAPIPort, "API server port")
	flag.StringVar(&cfg.path, "path", defaultPath, "Path to strong file")
	flag.StringVar(&cfg.stravaOAuth.ClientID, "strava-client", os.Getenv("STRAVA_CLIENT_ID"), "Strava API Client ID")
	flag.StringVar(&cfg.stravaOAuth.ClientSecret, "strava-secret", os.Getenv("STRAVA_CLIENT_SECRET"), "Strava API Client Secret")
	flag.StringVar(&cfg.stravaOAuth.RedirectURL, "redirect", defaultRedirectURL, "Strava Redirect URL")
	flag.StringVar(&cfg.gdriveOAuth.ClientID, "gdrive-client", os.Getenv("GDRIVE_CLIENT_ID"), "Google Drive API Client ID")
	flag.StringVar(&cfg.gdriveOAuth.ClientSecret, "gdrive-secret", os.Getenv("GDRIVE_CLIENT_SECRET"), "Google Drive API Client Secret")
	flag.Parse()

	if cfg.stravaOAuth.ClientID == "" {
		log.Print("strava client id is required")
		os.Exit(1)
	}

	if cfg.stravaOAuth.ClientSecret == "" {
		log.Print("strava client secret is required")
		os.Exit(1)
	}

	if cfg.gdriveOAuth.ClientID == "" {
		log.Print("gdrive client id is required")
		os.Exit(1)
	}

	if cfg.gdriveOAuth.ClientSecret == "" {
		log.Print("gdrive client secret is required")
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
	stravaClient := &strava.Provider{Logger: log, Config: cfg.stravaOAuth}

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
		log.Fatalf("error looking up user config directory %v", err)
	}

	filename := filepath.Join(path, "strong", "tokens.json")

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
