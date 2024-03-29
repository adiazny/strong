package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adiazny/strong/internal/pkg/auth"
	"github.com/adiazny/strong/internal/pkg/gdrive"
	"github.com/adiazny/strong/internal/pkg/store"
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

	gdriveTokenPath = "gdrive/storage.json"
	stravaTokenPath = "strava/storage.json"
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
	config             config
	log                *log.Logger
	stravaAuthProvider *auth.Provider
	gdriveAuthProvider *auth.Provider
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

	//========================================================================
	// Create files

	gStore, err := store.NewFile(gdriveTokenPath)
	if err != nil {
		log.Printf("error creating google drive file store %v\n", err)
		os.Exit(1)
	}

	stravaStore, err := store.NewFile(stravaTokenPath)
	if err != nil {
		log.Printf("error creating strava file store %v\n", err)
		os.Exit(1)
	}

	//========================================================================
	// Bootstrap OAuth Providers

	gdriveAuthProvider, err := auth.NewProvider(auth.GDriveService, gdriveTokenPath, cfg.gdriveClientID, cfg.gdriveClientSecret, cfg.gdriveRedirectURL, gStore)
	if err != nil {
		log.Printf("error creating gdrive auth provider %v\n", err)
		os.Exit(1)
	}

	stravaAuthProvider, err := auth.NewProvider(auth.StravaService, stravaTokenPath, cfg.stravaClientID, cfg.stravaClientSecret, cfg.stravaRedirectURL, stravaStore)
	if err != nil {
		log.Printf("error creating strava auth provider %v\n", err)
		os.Exit(1)
	}

	//========================================================================
	// API Server Flow

	app := &application{
		config:             cfg,
		log:                log,
		stravaAuthProvider: stravaAuthProvider,
		gdriveAuthProvider: gdriveAuthProvider,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("starting api server on %s", srv.Addr)
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("error encounted with http server %v\n", err)
		}

		log.Print("shutting down api server\n")
	}()

	//========================================================================
	// Google Drive Flow

	token, err := gdriveAuthProvider.Storage.GetToken()
	if err != nil || !token.Valid() {
		gdriveURL := gdriveAuthProvider.AuthCodeURL("gdrive-state")
		log.Println(gdriveURL)

		for gStore.FileNotPresent() {
			time.Sleep(10 * time.Second)
		}
	}

	gdriveHttpClient := gdriveAuthProvider.Client(ctx, token)

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(gdriveHttpClient))
	if err != nil {
		log.Fatalf("error to creating gdrive service: %v", err)
	}

	driveProvider := &gdrive.Provider{
		DataPath:     "strong.csv",
		DriveService: driveService,
	}

	driveBytes, err := driveProvider.Import(ctx)
	if err != nil {
		log.Fatalf("error to importing gdrive file: %v", err)
	}

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
	// Strava Flow

	token, err = stravaAuthProvider.Storage.GetToken()
	if err != nil || !token.Valid() {
		stravaURL := stravaAuthProvider.AuthCodeURL("strava-state")
		log.Println(stravaURL)

		// This should be the filestorage method
		for stravaStore.FileNotPresent() {
			time.Sleep(10 * time.Second)
		}
	}

	stravaClient := stravaAuthProvider.Client(ctx, token)

	stravaProvider := strava.NewProvider(log, stravaClient)

	err = stravaProvider.UploadNewWorkouts(context.Background(), workouts)
	if err != nil {
		log.Fatalf("error uploading strava activities %v", err)
	}

	log.Print("uploaded new workouts to strava")
}
