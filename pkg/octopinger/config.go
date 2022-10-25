package octopinger

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/katallaxie/pkg/utils/files"
	"golang.org/x/exp/maps"
)

// NodeList ...
type NodeList struct {
	nodes []string
}

// Nodes ...
func (n *NodeList) Nodes() []string {
	return n.nodes
}

// Load ...
func (n *NodeList) Load(base, file string) error {
	n.nodes = nil

	f, err := os.Open(path.Join(base, file))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		n.nodes = append(n.nodes, scanner.Text())
	}

	return nil
}

// NewNodeList ...
func NewNodeList() *NodeList {
	return &NodeList{}
}

// Config ...
type Config struct {
	ICMP ICMPConfig
}

// ICMPConfig ...
type ICMPConfig struct {
	Enabled  bool
	Timeout  time.Duration
	Interval time.Duration
	External []string
}

// Load ...
func (c ICMPConfig) Load(base string) (ICMPConfig, error) {
	exists, _ := files.FileExists(path.Join(base, "probes.icmp.enabled"))
	if !exists {
		return ICMPConfig{Enabled: false}, nil
	}

	defaults := map[string]string{
		"timeout":  "1m",
		"interval": "3s",
		"external": "",
	}

	cfg, err := loadKeyValues(base, "probes.icmp.properties")
	if err != nil {
		return ICMPConfig{}, err
	}

	maps.Copy(defaults, cfg)

	timeout, err := time.ParseDuration(defaults["timeout"])
	if err != nil {
		return ICMPConfig{}, err
	}

	interval, err := time.ParseDuration(defaults["interval"])
	if err != nil {
		return ICMPConfig{}, err
	}

	return ICMPConfig{
		Enabled:  true,
		Interval: interval,
		Timeout:  timeout,
	}, nil
}

// Load ...
func (c Config) Load(base string) (Config, error) {
	icmpCfg, err := ICMPConfig{}.Load(base)
	if err != nil {
		return Config{}, nil
	}

	return Config{
		ICMP: icmpCfg,
	}, nil
}

func loadKeyValues(base, file string) (map[string]string, error) {
	cfg := make(map[string]string)

	f, err := os.Open(path.Join(base, file))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) < 2 {
			continue
		}

		cfg[parts[0]] = parts[1]
	}

	return cfg, nil
}
