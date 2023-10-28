package data

import (
	"context"
)

type Importer interface {
	Import(ctx context.Context) ([]byte, error)
}
