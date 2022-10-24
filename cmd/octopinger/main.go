package main

import (
	"context"
	"errors"
	"time"

	"github.com/ionos-cloud/octopinger/pkg/octopinger"
	"github.com/katallaxie/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type flags struct {
	Debug     bool
	NodeList  string
	Interval  time.Duration
	Timeout   time.Duration
	ICMPProbe bool
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
	rootCmd.Flags().DurationVar(&f.Interval, "interval", time.Duration(time.Second*1), "interval")
	rootCmd.Flags().DurationVar(&f.Timeout, "timeout", time.Duration(time.Second*5), "timeout")
	rootCmd.Flags().StringVar(&f.NodeList, "config-nodes", "/etc/config/nodes", "node list")
	rootCmd.Flags().BoolVar(&f.ICMPProbe, "icmp-probe", true, "icmp")
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

	probes := make([]octopinger.Probe, 0)
	if f.ICMPProbe {
		probes = append(probes, octopinger.NewICMPProbe())
	}

	o := octopinger.NewServer(
		octopinger.WithNodeList(f.NodeList),
		octopinger.WithLogger(logger),
		octopinger.WithProbes(probes...),
		octopinger.WithInterval(f.Interval),
		octopinger.WithTimeout(f.Timeout),
	)
	srv.Listen(o, false)

	if err := srv.Wait(); errors.Is(err, &server.Error{}) {
		return err
	}

	return nil
}
