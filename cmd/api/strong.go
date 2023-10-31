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
	"github.com/adiazny/strong/internal/pkg/gdrive"
	"github.com/adiazny/strong/internal/pkg/strava"
	"github.com/adiazny/strong/internal/pkg/strong"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const version = "1.1.0"

const (
	defaultAPIPort     = 5000
	defaultPath        = "./strong.csv"
	defaultRedirectURL = "http://localhost:4001/v1/redirect"

	stravaType = 0
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
	config         config
	log            *log.Logger
	stravaProvider *strava.Provider
	gdriveProvider *gdrive.Provider
	workouts       []strong.Workout
}

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	ctx := context.Background()

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

	gdriveAuth, err := auth.NewProvider(auth.GDriveService, cfg.gdriveClientID, cfg.gdriveClientSecret, cfg.gdriveRedirectURL)
	if err != nil {
		log.Printf("error creating gdrive auth provider %v\n", err)
		os.Exit(1)
	}

	userHomePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error looking up user home directory %v", err)
	}

	gdrivePath := filepath.Join(userHomePath, "gdrive", "storage.json")
	var driveBytes []byte

	var driveProvider *gdrive.Provider

	driveAuthExists := gdriveAuth.Exists(gdrivePath)
	if driveAuthExists {
		gdriveHttpClient, err := gdriveAuth.HttpClient(ctx, gdrivePath)
		if err != nil {
			log.Fatalf("error creating gdrive http client %v", err)
		}

		driveService, err := drive.NewService(ctx, option.WithHTTPClient(gdriveHttpClient))
		if err != nil {
			log.Fatalf("error to creating gdrive service: %v", err)
		}

		driveProvider = &gdrive.Provider{
			Path:         gdrivePath,
			DriveService: driveService,
		}

		driveBytes, err = driveProvider.Import(context.Background())
		if err != nil {
			log.Fatalf("error to importing gdrive file: %v", err)
		}
	}

	if !driveAuthExists {
		// start a api server for google drive to send redirect
		// TODO: 10/30 continue with John here 
		gdriveURL := gdriveAuth.AuthCodeURL("gdriveState")
		log.Println(gdriveURL)
	}

	stravaAuth, err := auth.NewProvider(auth.StravaService, cfg.stravaClientID, cfg.stravaClientSecret, cfg.stravaRedirectURL)
	if err != nil {
		log.Printf("error creating strava auth provider %v\n", err)
		os.Exit(1)
	}

	//========================================================================
	// Local File Implementation
	// fp := &localData.FileProvider{
	// 	Path: cfg.path,
	// }

	// fileBytes, err := fp.Import(context.Background())
	// if err != nil {
	// 	log.Printf("error importing file %v\n", err)
	// 	os.Exit(1)
	// }
	// file := bytes.NewReader(fileBytes)

	var workouts []strong.Workout

	if driveBytes != nil {
		file := bytes.NewReader(driveBytes)

		workouts, err = strong.Process(file)
		if err != nil {
			log.Printf("error processing file %v\n", err)
			os.Exit(1)
		}
	} else {
		log.Printf("empty drive file imported\n")
		os.Exit(1)
	}

	//========================================================================
	// Strava bootstrap

	stravaProvider := strava.NewProvider(log, stravaAuth)

	//========================================================================
	// API Server Setup
	app := &application{
		config:         cfg,
		log:            log,
		stravaProvider: stravaProvider,
		gdriveProvider: driveProvider,
		workouts:       workouts,
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
	// OAuth Checks for Strava and Google Drive

	stravaPath := filepath.Join(userHomePath, "strava", "storage.json")

	stravaAuthExists := stravaAuth.Exists(stravaPath)
	if stravaAuthExists {
		tokens, err := stravaAuth.FileTokens(stravaPath)
		if err != nil {
			log.Fatalf("error getting strava tokens %v", err)
		}

		log.Print("uploading new workouts to strava")

		err = app.uploadNewWorkouts(context.Background(), tokens)
		if err != nil {
			log.Fatalf("error uploading strava activities %v", err)
		}

	} else {
		// Start api server if token file not found or errored during opening file
		go func() {
			log.Printf("starting api server on %s", srv.Addr)
			serverErrors <- srv.ListenAndServe()
		}()

		stravaURL := stravaAuth.AuthCodeURL("stravaState")
		log.Println(stravaURL)

		// Blocking main.
		if err := <-serverErrors; err != nil {
			log.Fatalf("error encounted with http server %v", err)
		}
	}

}
