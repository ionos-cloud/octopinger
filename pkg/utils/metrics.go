package utils

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// kubernetesDuration holds durations of the Kubernetes requests.
	kubernetesDuration *prometheus.HistogramVec
	once               sync.Once
)

// clientWithMetrics is a wrapper around Kubernetes client that proxies all the
// requests and updates corresponding metrics.
type clientWithMetrics struct {
	client client.Client
}

// NewClient creates a new Kubernetes Client that monitors Kubernetes operations.
func NewClient(client client.Client) client.Client {
	return &clientWithMetrics{client: client}
}

// DefaultNewClientWithMetrics supports read and list caching and enables
// monitoring measured by ignored cache requests.
// This method is used by a k8s manager.
func DefaultNewClientWithMetrics(
	cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object,
) (client.Client, error) {
	c, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	mc := NewClient(c)
	return client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader:     cache,
		Client:          mc,
		UncachedObjects: uncachedObjects,
	})
}

// CreateMetrics creates a new Kubernetes Histogram metric and registers it in
// the given RegistererGatherer using given namespace.
// It is safe to call that method multiple times, the metric will be created
// only once. Subsequent calls returns the same metric.
func CreateMetrics(namespace string) *prometheus.HistogramVec {
	once.Do(func() {
		kubernetesDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "k8s",
				Name:      "duration_seconds",
				Help:      "Duration of kubernetes requests partitioned by object kind and method",
			},
			[]string{"kind", "method"},
		)
	})

	return kubernetesDuration
}

// Status updates status subresource for k8s objects. It proxies requests to
// underlying controller-runtime StatusClient.
func (c *clientWithMetrics) Status() client.StatusWriter {
	return c.client.Status()
}

// Scheme returns the scheme this client is using. It proxies requests to
// underlying controller-runtime Client.
func (c *clientWithMetrics) Scheme() *runtime.Scheme {
	return c.client.Scheme()
}

// RESTMapper returns the rest this client is using. It proxies requests to
// underlying controller-runtime Client.
func (c *clientWithMetrics) RESTMapper() meta.RESTMapper {
	return c.client.RESTMapper()
}

// Get retrieves an obj for the given object key from the k8s Cluster. It
// proxies requests to underlying controller-runtime Reader.
func (c *clientWithMetrics) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return requestWithMeasure(
		"get",
		func() (runtime.Object, error) {
			err := c.client.Get(ctx, key, obj, opts...)
			return obj, err
		},
	)
}

// List retrieves list of objects for a given namespace and list options. It
// proxies requests to underlying controller-runtime Reader.
func (c *clientWithMetrics) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return requestWithMeasure(
		"list",
		func() (runtime.Object, error) {
			err := c.client.List(ctx, list, opts...)
			return list, err
		},
	)
}

// Create saves the object obj in the Kubernetes cluster. It proxies requests to
// underlying controller-runtime Writer.
func (c *clientWithMetrics) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return requestWithMeasure(
		"create",
		func() (runtime.Object, error) {
			err := c.client.Create(ctx, obj, opts...)
			return obj, err
		},
	)
}

// Delete deletes the given obj from Kubernetes cluster. It proxies requests to
// underlying controller-runtime Writer.
func (c *clientWithMetrics) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return requestWithMeasure(
		"delete",
		func() (runtime.Object, error) {
			return obj, c.client.Delete(ctx, obj, opts...)
		},
	)
}

// Update updates the given obj in the Kubernetes cluster. It proxies requests to
// underlying controller-runtime Writer.
func (c *clientWithMetrics) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return requestWithMeasure(
		"update",
		func() (runtime.Object, error) {
			return obj, c.client.Update(ctx, obj, opts...)
		},
	)
}

// Patch patches the given obj in the Kubernetes cluster. It proxies requests to
// underlying controller-runtime Writer.
func (c *clientWithMetrics) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return requestWithMeasure(
		"patch",
		func() (runtime.Object, error) {
			return obj, c.client.Patch(ctx, obj, patch, opts...)
		},
	)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
// It proxies requests to underlying controller-runtime Writer.
func (c *clientWithMetrics) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return requestWithMeasure(
		"delete_all_of",
		func() (runtime.Object, error) {
			return obj, c.client.DeleteAllOf(ctx, obj, opts...)
		},
	)
}

// requestWithMeasure executes the given function f and updates corresponding
// Kubernetes metrics.
func requestWithMeasure(method string, f func() (runtime.Object, error)) error {
	start := time.Now()

	obj, err := f()

	duration := time.Since(start)
	kind := GetKind(obj)
	if kubernetesDuration != nil {
		kubernetesDuration.WithLabelValues(kind, method).Observe(duration.Seconds())
	}

	return err
}
