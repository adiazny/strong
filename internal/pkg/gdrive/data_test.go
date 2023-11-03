// Package gdrive will handle downloading the latest strong csv file
// Reference testing-code-that-depends-on-googlegolangorgapi https://github.com/googleapis/google-api-go-client/blob/v0.149.0/testing.md
package gdrive_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/adiazny/strong/internal/pkg/gdrive"
)

func TestProvider_Import(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		p       *gdrive.Provider
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.p.Import(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Import() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.Import() = %v, want %v", got, tt.want)
			}
		})
	}
}
