// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"testing"
)

func Test_manifestService_Validate(t *testing.T) {
	type fields struct {
		filePath string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantErr           bool
		wantValidationErr bool
	}{
		{
			name: "read the manifest from the file",
			fields: fields{
				filePath: "test/manifest_1/manifest.json",
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:           false,
			wantValidationErr: false,
		},
		{
			name: "read the manifest with wrong extension name from the file",
			fields: fields{
				filePath: "test/manifest_1/manifest_with_err.json",
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:           false,
			wantValidationErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			manifestLoader, _ := LoaderFactory(tt.fields.filePath)
			m, err := NewManifestService(ctx, manifestLoader)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManifestService error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := m.Validate(); (err != nil) != tt.wantValidationErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantValidationErr)
			}
		})
	}
}
