package data

import (
	"context"
	"io"
	"os"
)

type FileProvider struct {
	FilePath string
}

func (fp *FileProvider) Import(ctx context.Context) ([]byte, error) {
	file, err := os.Open(fp.FilePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
