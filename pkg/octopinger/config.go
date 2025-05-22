package octopinger

import (
	"bufio"
	"encoding/json"
	"os"
	"path"

	"github.com/ionos-cloud/octopinger/api/v1alpha1"
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
		defer func() {
		    _ = f.Close()
		}()

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

	for _, filter := range n.filters {
		for i := len(nodes) - 1; i >= 0; i-- {
			if filter(nodes[i]) {
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

type config struct{}

// Config ...
func Config() config {
	return config{}
}

// Load ...
func (c config) Load(base string) (*v1alpha1.Config, error) {
	p := path.Clean(path.Join(base, "config"))

	file, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	cfg := &v1alpha1.Config{}
	err = json.Unmarshal(file, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
