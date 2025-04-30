package source

import (
	"testing"
)

func TestLocalSource_ResolveSourcePath(t *testing.T) {
	type fields struct {
		LocalPath    string
		ManifetsPath string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Local path is empty",
			fields: fields{
				LocalPath:    "",
				ManifetsPath: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/examples/manifest.json",
			},
			want: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/examples",
		},
		{
			name: "Manifest path is empty",
			fields: fields{
				LocalPath:    "../../elements/agent",
				ManifetsPath: "",
			},
			want: "../../elements/agent",
		},
		{
			name: "Local path is one level up",
			fields: fields{
				LocalPath:    "../elements/agent",
				ManifetsPath: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/examples/manifest.json",
			},
			want: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/elements/agent",
		},
		{
			name: "Local path is two levels up",
			fields: fields{
				LocalPath:    "../../elements/agent",
				ManifetsPath: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/examples/manifest.json",
			},
			want: "/Users/lpuskas/prj/cisco-eti/elements/agent",
		},
		{
			name: "Local path is absolute",
			fields: fields{
				LocalPath:    "/elements/agent",
				ManifetsPath: "/Users/lpuskas/prj/cisco-eti/agent-workflow-cli/examples/manifest.json",
			},
			want: "/elements/agent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LocalSource{
				LocalPath:    tt.fields.LocalPath,
				ManifestPath: tt.fields.ManifetsPath,
			}
			if got := ls.ResolveSourcePath(); got != tt.want {
				t.Errorf("ResolveSourcePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
