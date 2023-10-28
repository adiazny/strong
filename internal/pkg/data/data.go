package data

import (
	"context"
	"os"
)

type Importer interface {
	Import(ctx context.Context) (*os.File, error)
}
