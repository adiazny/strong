package data

import (
	"context"
	"io"
	"os"
)

type FileProvider struct {
	Path string
}

func (fp *FileProvider) Import(ctx context.Context) ([]byte, error) {
	file, err := os.Open(fp.Path)
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
