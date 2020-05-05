package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	zapLogger "github.com/hatappi/go-recmd/internal/logger/zap"
)

var (
	verbose  bool
	logLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "recmd",
	Short:        "recmd is live reloading tool for any application",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := zapLogger.FromContext(ctx)

		if verbose {
			logLevel.SetLevel(zapcore.DebugLevel)
			logger.Debug("enable verbose mode")
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx := context.Background()

	logger, err := zapLogger.NewZap(logLevel)
	if err != nil {
		fmt.Printf("logger initialize error: %v", err)
		os.Exit(1)
	}
	ctx = zapLogger.WithContext(ctx, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error("recmd execute failed", zap.Error(err))
	}
}

func init() {
	rootCmd.AddCommand(newWatchCmd())

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "enable verbose")
}
