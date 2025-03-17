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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "read the manifest from the file",
			fields: fields{
				filePath: "../../../../examples/manifest.json",
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := manifestService{
				filePath: tt.fields.filePath,
			}
			if err := m.Validate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
