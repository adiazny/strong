// Package gdrive will handle downloading the latest strong file
package gdrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"google.golang.org/api/drive/v3"
)

const (
	createdTimeDescending = "createdTime desc"
	driveFilesPageSize    = 5
)

// GOAL:
// Search latest strong.csv file
// Download strong.csv file

type FileProvider struct {
	//logger
	Path         string
	DriveService *drive.Service
}

func (fp *FileProvider) Import(ctx context.Context) ([]byte, error) {
	fileName := path.Base(fp.Path)

	driveFile, err := fp.searchLatest(fileName)
	if err != nil {
		return nil, fmt.Errorf("error searching drive file %s", fileName)
	}

	if driveFile.Id == "" {
		return nil, fmt.Errorf("error file ID is empty for %s", fileName)
	}

	fileBytes, err := fp.download(driveFile.Id)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (fp *FileProvider) searchLatest(fileName string) (*drive.File, error) {
	query := fmt.Sprintf("name = '%s'", fileName)

	fileListCall, err := fp.DriveService.Files.List().PageSize(driveFilesPageSize).OrderBy(createdTimeDescending).Q(query).Do()
	if err != nil {
		return nil, err
	}

	if len(fileListCall.Files) == 0 {
		return nil, errors.New("no google drive files found")
	}

	return fileListCall.Files[0], nil
}

func (fp *FileProvider) download(fileId string) ([]byte, error) {
	response, err := fp.DriveService.Files.Get(fileId).Download()
	if err != nil {
		return nil, fmt.Errorf("error downloading file %s: %w", fileId, err)
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading drive body: %w", err)
	}

	return bytes, nil
}
