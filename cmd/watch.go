package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	zapLogger "github.com/hatappi/go-recmd/internal/logger/zap"
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
				"watch command options",
				zap.String("path", opts.path),
				zap.Any("exclude", opts.excludes),
				zap.Any("commands", args),
			)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.path, "path", "p", opts.path, "watch path")
	cmd.Flags().StringSliceVarP(&opts.excludes, "exclude", "e", opts.excludes, "exclude path. you can specify multiple")
	cmd.Flags().SetInterspersed(false)

	return cmd
}
