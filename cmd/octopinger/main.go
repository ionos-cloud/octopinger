package main

import (
	"context"
	"errors"

	"github.com/ionos-cloud/octopinger/pkg/octopinger"
	"github.com/katallaxie/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type flags struct {
	Debug      bool
	ConfigPath string
}

var f = &flags{}

var rootCmd = &cobra.Command{
	Use: "octopinger",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context())
	},
}

func init() {
	rootCmd.Flags().BoolVar(&f.Debug, "debug", f.Debug, "debug")
	rootCmd.Flags().StringVar(&f.ConfigPath, "config", "/etc/config", "config")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	defer func() { _ = logger.Sync() }()

	logger.Info("Octopinger")

	srv, _ := server.WithContext(ctx)

	api := octopinger.NewAPI()
	srv.Listen(api, false)

	o := octopinger.NewServer(
		octopinger.WithLogger(logger),
		octopinger.WithConfigPath(f.ConfigPath),
	)
	srv.Listen(o, false)

	if err := srv.Wait(); errors.Is(err, &server.Error{}) {
		return err
	}

	return nil
}
