package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"
)

// File satisfies the auth.Storage interface
type File struct {
	path string
	mu   sync.RWMutex
}

func NewFile(path string) (*File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(homeDir, path)

	return &File{path: fullPath}, nil
}

// GetToken retrieves a token from a file
func (f *File) GetToken() (*oauth2.Token, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	in, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	var t *oauth2.Token

	data := json.NewDecoder(in)

	return t, data.Decode(&t)
}

// SetToken creates, truncates, then stores a token
// in a file
func (f *File) SetToken(t *oauth2.Token) error {
	if t == nil || !t.Valid() {
		return errors.New("error bad token")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	out, err := os.OpenFile(f.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer out.Close()

	data, err := json.Marshal(&t)
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}

func (f *File) TokenNotPresent() bool {
	_, err := os.Stat(f.path)
	return err != nil
}
