package watcher

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIsWatchDir(t *testing.T) {
	os.MkdirAll("tmp/test/a/b/c/d", 0755)
	defer os.RemoveAll("tmp")

	testCases := []struct {
		path      string
		targetDir string
		expect    bool
		wantErr   bool
	}{
		{
			path:      "**/*",
			targetDir: "tmp/test/a",
			expect:    true,
			wantErr:   false,
		},
		{
			path:      "tmp/test/*",
			targetDir: "tmp/test",
			expect:    true,
			wantErr:   false,
		},
		{
			path:      "tmp/**/c/d/",
			targetDir: "tmp/test/a/b/c/d",
			expect:    true,
			wantErr:   false,
		},
		{
			path:      "tmp", // syntax sugar
			targetDir: "tmp/test/a",
			expect:    true,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path is %s, targetDir is %s", tc.path, tc.targetDir), func(tt *testing.T) {
			w, err := NewWatcher(tc.path, []string{}, nil, nil)
			if err != nil {
				tt.Fatal(err)
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
	defer os.RemoveAll("tmp")

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
			path: "tmp", // syntax sugar
			expect: []string{
				"tmp",
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
				"tmp/test/a/b",
			},
			wantErr: false,
		},
		{
			path: "tmp/test/*/*/c/d/",
			expect: []string{
				"tmp/test/a/b/c/d",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(tt *testing.T) {
			w, err := NewWatcher(tc.path, []string{}, nil, nil)
			if err != nil {
				tt.Fatal(err)
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

func TestConvertRegexp(t *testing.T) {
	os.MkdirAll("tmp/test/a/b/c/d", 0755)
	defer os.RemoveAll("tmp")

	testCases := []struct {
		path    string
		expect  string
		wantErr bool
	}{
		{
			path:    "tmp/**/*",
			expect:  "^tmp/([^/]*/)*$",
			wantErr: false,
		},
		{
			path:    "tmp/",
			expect:  "^tmp/([^/]*/)*([^/]*/)?$",
			wantErr: false,
		},
		{
			path:    "tmp/test",
			expect:  "^tmp/test/$",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		actual, err := convertRegexp(tc.path)

		if (err == nil) == tc.wantErr {
			t.Fatalf("err unexpected, %+v", err)
		}

		if tc.expect != actual.String() {
			t.Fatalf("unexpected regepx expect: %s, actual: %s", tc.expect, actual.String())
		}
	}
}

func TestNormalizePath(t *testing.T) {
	os.MkdirAll("tmp/test/a/b/c/d", 0755)
	defer os.RemoveAll("tmp")

	testCases := []struct {
		path   string
		expect string
	}{
		{
			path:   "tmp/test/",
			expect: "tmp/test/",
		},
		{
			path:   "tmp/test",
			expect: "tmp/test/",
		},
		{
			path:   "tmp/test/**/*",
			expect: "tmp/test/**/*",
		},
		{
			path:   "./tmp",
			expect: "tmp/",
		},
	}

	for _, tc := range testCases {
		actual := normalizePath(tc.path)

		if tc.expect != actual {
			t.Fatalf("unexpected path. expect: %s, actual: %s", tc.expect, actual)
		}
	}
}
