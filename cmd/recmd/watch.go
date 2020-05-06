package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/hatappi/go-recmd/internal/event"
	"github.com/hatappi/go-recmd/internal/executor"
	zapLogger "github.com/hatappi/go-recmd/internal/logger/zap"
	"github.com/hatappi/go-recmd/internal/watcher"
)

var watchCmdExample = `
$ recmd watch go run main.go
$ recmd watch -p "./main.go" go run main.go
$ recmd watch --exclude testA --exclude testB go run main.go
`

type watchOption struct {
	path     string
	excludes []string
}

func newWatchCmd() *cobra.Command {
	opts := &watchOption{
		path: "**/*",
	}

	cmd := &cobra.Command{
		Use:     "watch [flags] [your command]",
		Short:   "watch path and execute command",
		Aliases: []string{"w"},
		Example: watchCmdExample,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := zapLogger.FromContext(ctx)

			logger.Debug(
				"command options",
				zap.String("path", opts.path),
				zap.Strings("exclude", opts.excludes),
				zap.Strings("commands", args),
			)

			eventChan := make(chan *event.Event)

			eg := errgroup.Group{}

			w, err := watcher.NewWatcher(opts.path, opts.excludes, eventChan, logger)
			if err != nil {
				return err
			}

			executor := executor.NewExecutor(logger, eventChan)

			ctx, cancel := context.WithCancel(ctx)

			eg.Go(func() error {
				defer cancel()
				return w.Run(ctx)
			})

			eg.Go(func() error {
				defer cancel()
				return executor.Run(ctx, args)
			})

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				for range c {
					cancel()
				}
			}()

			if err := eg.Wait(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.path, "path", "p", opts.path, "watch path")
	cmd.Flags().StringSliceVarP(&opts.excludes, "exclude", "e", opts.excludes, "exclude path. you can specify multiple it")
	cmd.Flags().SetInterspersed(false)

	return cmd
}
