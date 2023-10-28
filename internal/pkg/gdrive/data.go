// Package gdrive will handle downloading the latest strong file
package gdrive

import (
	"context"
	"os"
)

type FileProvider struct{}

func (fp *FileProvider) Import(ctx context.Context) (*os.File, error) {
	// perform google drive api request to download file
	return nil, nil
}
