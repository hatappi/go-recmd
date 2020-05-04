package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	zapLogger "github.com/hatappi/go-recmd/internal/logger/zap"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:     "watch [flags] [your command]",
	Short:   "watch path and execute command",
	Aliases: []string{"w"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := zapLogger.FromContext(ctx)

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}

		excludes, err := cmd.Flags().GetStringSlice("exclude")
		if err != nil {
			return err
		}

		logger.Error(
			"watch command options",
			zap.String("path", path),
			zap.Any("exclude", excludes),
			zap.Any("commands", args),
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().StringP("path", "p", "**/*", "watch path")
	watchCmd.Flags().StringSlice("exclude", []string{}, "exclude path. you can specify multiple")

	watchCmd.Flags().SetInterspersed(false)
}
