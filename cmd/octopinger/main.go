package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/ionos-cloud/octopinger/pkg/octopinger"
	"github.com/katallaxie/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var build = fmt.Sprintf("%s (%s) (%s)", version, commit, date)

type flags struct {
	Debug      bool
	ConfigPath string `env:"CONFIG_PATH" envDefault:"/etc/config"`
	StatusAddr string `env:"STATUS_ADDR" envDefault:"0.0.0.0:8081"`
	PodIP      string `env:"POD_IP"`
	HostIP     string `env:"HOST_IP"`
	Nodename   string `env:"NODE_NAME"`
}

var f = &flags{}

var rootCmd = &cobra.Command{
	Use:     "octopinger",
	Version: build,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runE(cmd.Context())
	},
}

func init() {
	if err := env.Parse(f); err != nil {
		panic(err)
	}

	rootCmd.Flags().BoolVar(&f.Debug, "debug", f.Debug, "debug")
	rootCmd.Flags().StringVar(&f.ConfigPath, "config", f.ConfigPath, "config")
	rootCmd.Flags().StringVar(&f.StatusAddr, "status-addr", f.StatusAddr, "status addr")
	rootCmd.Flags().StringVar(&f.Nodename, "nodename", f.Nodename, "node name")
	rootCmd.Flags().StringVar(&f.PodIP, "pod-ip", f.PodIP, "pod ip")
	rootCmd.Flags().StringVar(&f.HostIP, "host-ip", f.HostIP, "host ip")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func runE(ctx context.Context) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	defer func() { _ = logger.Sync() }()

	logger.Sugar().Infow("starting octopinger", "build", build, "nodename", f.Nodename, "pod-ip", f.PodIP, "host-ip", f.HostIP)

	srv, _ := server.WithContext(ctx)

	err = octopinger.DefaultRegisterer.Register(octopinger.DefaultMetrics)
	if err != nil {
		return err
	}

	m := octopinger.NewMonitor(octopinger.DefaultMetrics)

	api := octopinger.NewAPI(
		octopinger.WithAddr(f.StatusAddr),
	)
	srv.Listen(api, false)

	o := octopinger.NewServer(
		octopinger.WithLogger(logger),
		octopinger.WithConfigPath(f.ConfigPath),
		octopinger.WithMonitor(m),
		octopinger.WithNodeName(f.Nodename),
		octopinger.WithPodIP(f.PodIP),
		octopinger.WithHostIP(f.HostIP),
	)
	srv.Listen(o, false)

	if err := srv.Wait(); errors.Is(err, &server.Error{}) {
		return err
	}

	return nil
}
