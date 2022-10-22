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
	Debug bool
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

	defer logger.Sync()

	logger.Info("Goldpinger")

	srv, _ := server.WithContext(ctx)

	o := octopinger.NewServer()
	srv.Listen(o, false)

	if err := srv.Wait(); errors.Is(err, &server.Error{}) {
		return err
	}

	return nil
}
