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

type Provider struct {
	//logger
	DataPath     string
	DriveService *drive.Service
}

func (p *Provider) Import(ctx context.Context) ([]byte, error) {
	fileName := path.Base(p.DataPath)

	driveFile, err := p.searchLatest(fileName)
	if err != nil {
		return nil, fmt.Errorf("error searching drive file %s: %w", fileName, err)
	}

	if driveFile.Id == "" {
		return nil, fmt.Errorf("error file ID is empty for %s", fileName)
	}

	fileBytes, err := p.download(driveFile.Id)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (p *Provider) searchLatest(fileName string) (*drive.File, error) {
	query := fmt.Sprintf("name = '%s'", fileName)

	fileListCall, err := p.DriveService.Files.List().PageSize(driveFilesPageSize).OrderBy(createdTimeDescending).Q(query).Do()
	if err != nil {
		return nil, err
	}

	if len(fileListCall.Files) == 0 {
		return nil, errors.New("no google drive files found")
	}

	return fileListCall.Files[0], nil
}

func (p *Provider) download(fileId string) ([]byte, error) {
	response, err := p.DriveService.Files.Get(fileId).Download()
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
