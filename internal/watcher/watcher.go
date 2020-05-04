// Package watcher watch change event
package watcher

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	e "github.com/hatappi/go-recmd/internal/event"
	zapLogger "github.com/hatappi/go-recmd/internal/logger/zap"
)

// Watcher represent watcher interface
type Watcher interface {
	Run(ctx context.Context) error
}

type watcher struct {
	path      string
	eventChan chan *e.Event
}

// NewWatcher initilize watcher
func NewWatcher(path string, eventChan chan *e.Event) Watcher {
	return &watcher{
		path:      filepath.Clean(path),
		eventChan: eventChan,
	}
}

func (w *watcher) Run(ctx context.Context) error {
	logger := zapLogger.FromContext(ctx)

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
				logger.Debug("occur event", zap.Any("event", event), zap.Bool("ok", ok))
				if !ok {
					continue
				}

				if event.Op != fsnotify.Remove {
					fileInfo, osErr := os.Stat(event.Name)
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
								logger.Error("directory watch is failed", zap.String("path", newDir), zap.Error(err))
								return
							}
							if isWatchDir {
								logger.Debug("watch add new directory", zap.String("path", newDir))
								if err = watcher.Add(newDir); err != nil {
									logger.Error("directory add is failed", zap.String("path", newDir), zap.Error(err))
									return
								}
							}
						}(event.Name)
					}
				}

				_, f := filepath.Split(event.Name)
				if fileMatchPattern.MatchString(f) {
					w.eventChan <- &e.Event{
						Path: event.Name,
					}
				}
			case watchErr, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				logger.Error("watch is failed", zap.Error(watchErr))
			}
		}
	})

	for _, wd := range watchDir {
		logger.Debug("watch directory", zap.String("path", wd))
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

	rootPath := strings.Split(w.path, "/")[0]
	if strings.Contains(rootPath, "*") {
		rootPath = "."
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
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
	if strings.HasPrefix(targetDir, ".git") {
		return false, nil
	}

	dir, _ := filepath.Split(w.path)
	dir = strings.TrimRight(dir, "/")

	splitDir := strings.Split(dir, "/")
	patterns := []string{}

	for _, d := range splitDir {
		if d == "**" {
			patterns = append(patterns, "([^/]*/)*")
			continue
		}

		if strings.Contains(d, "*") {
			d = strings.Replace(d, "*", "([^/]*)?", -1)
		}

		patterns = append(patterns, d+"/")
	}

	pattern := strings.Join(patterns, "")
	pattern = "^" + pattern + "$"
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	targetDir = strings.TrimRight(targetDir, "/")
	targetDir += "/"
	return r.MatchString(targetDir), nil
}
