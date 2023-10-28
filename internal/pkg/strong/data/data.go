package data

import (
	"context"
	"os"
)

type FileProvider struct {
	Path string
}

func (fp *FileProvider) Import(ctx context.Context) (*os.File, error) {
	file, err := os.Open(fp.Path)
	if err != nil {
		return nil, err
	}

	return file, nil
}
