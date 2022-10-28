package utils

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsObjectFound ...
func IsObjectFound(ctx context.Context, client client.Client, namespace string, name string, obj client.Object) bool {
	return !apierrors.IsNotFound(FetchObject(ctx, client, namespace, name, obj))
}

// FetchObject ...
func FetchObject(ctx context.Context, client client.Client, namespace string, name string, obj client.Object) error {
	return client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
}
