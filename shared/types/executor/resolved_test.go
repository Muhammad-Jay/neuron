package executor

import (
	"path/filepath"
	"testing"
)

func TestEntrypointPath(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "store", "github", "read", "1.2.0")

	tests := []struct {
		name       string
		rootDir    string
		entrypoint string
		want       string
	}{
		{
			name:       "nested entrypoint uses native separators",
			rootDir:    root,
			entrypoint: "bin/executor",
			want:       filepath.Join(root, "bin", "executor"),
		},
		{
			name:       "empty root returns entrypoint unchanged",
			rootDir:    "",
			entrypoint: "executor.sh",
			want:       "executor.sh",
		},
		{
			name:       "empty entrypoint returns empty",
			rootDir:    root,
			entrypoint: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResolvedExecutor{
				RootDir: tt.rootDir,
				Runtime: RuntimeInfo{Entrypoint: tt.entrypoint},
			}
			if got := r.EntrypointPath(); got != tt.want {
				t.Errorf("EntrypointPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
