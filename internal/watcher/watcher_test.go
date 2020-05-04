package watcher

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIsWatchDir(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		targetDir string
		expect    bool
		wantErr   bool
	}{
		{
			name:      "path is **/*, targetDir is test/test/example",
			path:      "**/*",
			targetDir: "test/test/example",
			expect:    true,
			wantErr:   false,
		},
		{
			name:      "path is test/*, targetDir is test",
			path:      "test/*",
			targetDir: "test2/test",
			expect:    false,
			wantErr:   false,
		},
		{
			name:      "path is test/*/*/test, targetDir is test/a/b",
			path:      "test/*/*/test",
			targetDir: "test/a/b",
			expect:    true,
			wantErr:   false,
		},
		{
			name:      "path is test/**/example/test, targetDir is test/a/b/example",
			path:      "test/**/example/test",
			targetDir: "test/a/b/example",
			expect:    true,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			w := &watcher{
				path: tc.path,
			}

			actual, err := w.isWatchDir(tc.targetDir)

			if (err == nil) == tc.wantErr {
				tt.Fatalf("err unexpected, %+v", err)
			}

			if actual != tc.expect {
				tt.Fatalf("unexpected value. actual: %v, expect: %v", actual, tc.expect)
			}
		})
	}
}

func TestGetWatchDirs(t *testing.T) {
	os.MkdirAll("tmp/test/a/b/c/d", 0755)
	defer os.Remove("tmp")

	testCases := []struct {
		path    string
		expect  []string
		wantErr bool
	}{
		{
			path: "tmp/test/**/*",
			expect: []string{
				"tmp/test",
				"tmp/test/a",
				"tmp/test/a/b",
				"tmp/test/a/b/c",
				"tmp/test/a/b/c/d",
			},
			wantErr: false,
		},
		{
			path: "tmp/test/a/b",
			expect: []string{
				"tmp/test/a",
			},
			wantErr: false,
		},
		{
			path: "tmp/test/*/*/c/d",
			expect: []string{
				"tmp/test/a/b/c",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(tt *testing.T) {
			w := &watcher{
				path: tc.path,
			}

			actual, err := w.getWatchDirs()

			if (err == nil) == tc.wantErr {
				tt.Fatalf("err unexpected, %+v", err)
			}

			if diff := cmp.Diff(tc.expect, actual); diff != "" {
				t.Fatalf("getWatchDirs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
