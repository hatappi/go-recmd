// Package watcher watch change event
package watcher

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	e "github.com/hatappi/go-recmd/internal/event"
)

var defaultExcludePaths = []string{
	".git/**/*",
}

// Watcher represent watcher interface
type Watcher interface {
	Run(ctx context.Context) error

	isWatchDir(targetDir string) (bool, error)
	getWatchDirs() ([]string, error)
}

type watcher struct {
	path            string
	pathPattern     *regexp.Regexp
	excludePaths    []string
	excludePatterns []*regexp.Regexp
	eventChan       chan *e.Event
	logger          *zap.Logger
}

// NewWatcher initilize watcher
func NewWatcher(path string, excludePaths []string, eventChan chan *e.Event, logger *zap.Logger) (Watcher, error) {
	neps := []string{}
	eps := []*regexp.Regexp{}
	for _, p := range append(defaultExcludePaths, excludePaths...) {
		n := normalizePath(p)
		neps = append(neps, n)
		ep, err := convertRegexp(n)
		if err != nil {
			return nil, err
		}
		eps = append(eps, ep)
	}

	np := normalizePath(path)
	pp, err := convertRegexp(np)
	if err != nil {
		return nil, err
	}

	return &watcher{
		path:            np,
		pathPattern:     pp,
		excludePaths:    neps,
		excludePatterns: eps,
		eventChan:       eventChan,
		logger:          logger,
	}, nil
}

func (w *watcher) Run(ctx context.Context) error {
	watchDir, err := w.getWatchDirs()
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		_ = watcher.Close()
	}()

	_, filename := filepath.Split(w.path)
	rep, err := regexp.Compile(`\*+`)
	if err != nil {
		return err
	}
	fileMatchPattern, err := regexp.Compile(rep.ReplaceAllString(filename, ".*") + "$")
	if err != nil {
		return err
	}

	eg := errgroup.Group{}
	eg.Go(func() error {
		for {
			select {
			case event, ok := <-watcher.Events:
				w.logger.Debug("occur event", zap.Reflect("event", event), zap.Bool("ok", ok))
				if !ok {
					continue
				}

				newPath := normalizePath(event.Name)

				if event.Op != fsnotify.Remove {
					fileInfo, osErr := os.Stat(newPath)
					if osErr != nil {
						// If temporary file is checked, it may not be visible.
						if _, ok := osErr.(*os.PathError); ok {
							continue
						}
						return err
					}

					if fileInfo.IsDir() {
						go func(newDir string) {
							var isWatchDir bool
							if isWatchDir, err = w.isWatchDir(newDir); err != nil {
								w.logger.Warn("failed to judge the watch target", zap.String("path", newDir), zap.Error(err))
								return
							}
							if isWatchDir {
								w.logger.Debug("add new directory to watch target", zap.String("path", newDir))
								if err = watcher.Add(newDir); err != nil {
									w.logger.Warn("failed to watch new directory", zap.String("path", newDir), zap.Error(err))
									return
								}
							}
						}(newPath)
					}
				}

				_, f := filepath.Split(newPath)
				if fileMatchPattern.MatchString(f) {
					w.eventChan <- &e.Event{
						Path:      newPath,
						CreatedAt: time.Now(),
					}
				}
			case watchErr, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				w.logger.Warn("failed to watch", zap.Error(watchErr))
			case <-ctx.Done():
				w.logger.Debug("finish watcher")
				return nil
			}
		}
	})

	for _, wd := range watchDir {
		w.logger.Debug("add directory to watch target", zap.String("path", wd))
		err = watcher.Add(wd)
		if err != nil {
			return err
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (w *watcher) getWatchDirs() ([]string, error) {
	watchDirs := []string{}

	r, err := regexp.Compile("^(/?[^/]+)")
	if err != nil {
		return nil, err
	}
	rootPath := r.FindString(w.path)
	if strings.Contains(rootPath, "*") {
		rootPath = "."
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		isPathMatch, err := w.isWatchDir(path)
		if err != nil {
			return err
		}
		if isPathMatch {
			watchDirs = append(watchDirs, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return watchDirs, nil
}

func (w *watcher) isWatchDir(targetDir string) (bool, error) {
	targetDir = normalizePath(targetDir)

	for _, r := range w.excludePatterns {
		if r.MatchString(targetDir) {
			return false, nil
		}
	}

	return w.pathPattern.MatchString(targetDir), nil
}

func convertRegexp(path string) (*regexp.Regexp, error) {
	path = normalizePath(path)
	dir, _ := filepath.Split(path)

	dir = strings.TrimRight(dir, "/")

	// syntax sugar
	r, err := regexp.Compile("^/?[^/]+$")
	if err != nil {
		return nil, err
	}
	if r.MatchString(dir) {
		dir += "/**/*"
	}

	splitDir := strings.Split(dir, "/")

	patterns := []string{}
	for _, d := range splitDir {
		// ignore current directory
		if d == "." {
			continue
		}

		if d == "**" {
			patterns = append(patterns, "([^/]*/)*")
			continue
		}

		if d == "*" {
			patterns = append(patterns, "([^/]*/)?")
			continue
		}

		if strings.Contains(d, "*") {
			d = strings.Replace(d, "*", "([^/]*)?", -1)
		}

		patterns = append(patterns, d+"/")
	}

	pattern := strings.Join(patterns, "")
	pattern = "^" + pattern + "$"
	r, err = regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func normalizePath(path string) string {
	// if path is glob pattern, it doesn't normalize
	if strings.Contains(path, "*") {
		return path
	}

	path = filepath.Clean(path)

	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		path = strings.TrimRight(path, "/")
		path += "/"
	}

	return path
}
