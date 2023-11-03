package data_test

import (
	"context"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/adiazny/strong/internal/pkg/strong/data"
)

func TestFileProvider_Import(t *testing.T) {

	tests := []struct {
		name           string
		strongFilePath string
		want           []byte
		wantErr        bool
	}{
		{
			name:           "success",
			strongFilePath: "testdata/sample_strong.csv",
			want:           wantBytes(t, "testdata/sample_strong.csv"),
			wantErr:        false,
		},
		{
			name:           "fail opening file",
			strongFilePath: "testdata/does_not_exist.csv",
			wantErr:        true,
		},
		{
			name:           "fail opening file",
			strongFilePath: "testdata/does_not_exist.csv",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fp := &data.FileProvider{
				FilePath: tt.strongFilePath,
			}
			got, err := fp.Import(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("FileProvider.Import() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileProvider.Import() = %v, want %v", got, tt.want)
			}
		})
	}
}

func wantBytes(t *testing.T, filePath string) []byte {
	t.Helper()

	bytes, err := os.ReadFile(path.Clean(filePath))
	if err != nil {
		t.Fatalf("error reading file %s: %v", filePath, err)
	}

	return bytes
}
