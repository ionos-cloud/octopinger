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

// NodeFilter ...
type NodeFilter func(node string) bool

// FilterIP ...
func FilterIP(ip string) NodeFilter {
	return func(node string) bool {
		return node == ip
	}
}

// NodeLoader ...
type NodeLoader func() ([]string, error)

// NodeLoader ...
func NodesLoader(base string) NodeLoader {
	return func() ([]string, error) {
		p := path.Clean(path.Join(base, "nodes"))
		nodes := make([]string, 0)

		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			nodes = append(nodes, scanner.Text())
		}

		return nodes, nil
	}
}

// Load ...
func (n *NodeList) Load() ([]string, error) {
	nodes := make([]string, 0)

	for _, loader := range n.loaders {
		n, err := loader()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n...)
	}

	for i := len(nodes) - 1; i >= 0; i-- {
		for _, filter := range n.filters {
			ok := filter(nodes[i])
			if ok {
				nodes = append(nodes[:i], nodes[i+1:]...)
			}
		}
	}

	return nodes, nil
}

// NodeList ...
type NodeList struct {
	loaders []NodeLoader
	filters []NodeFilter
}

// NewNodeList ...
func NewNodeList(loaders []NodeLoader, filters ...NodeFilter) *NodeList {
	n := new(NodeList)
	n.loaders = loaders
	n.filters = filters

	return n
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
