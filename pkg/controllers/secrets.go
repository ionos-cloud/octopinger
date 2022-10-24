package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NewSecretsReconciler ...
func NewSecretsReconciler(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Owns(&corev1.Pod{}).
		Complete(&secretReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
}

type secretReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile ...
func (s *secretReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	ds := &appsv1.DaemonSet{}
	err := s.Get(ctx, r.NamespacedName, ds)
	if err != nil {
		return reconcile.Result{}, nil
	}

	pods := &corev1.PodList{}
	err = s.List(ctx, pods, client.InNamespace(r.Namespace), client.MatchingLabels(ds.Spec.Template.Labels))
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, p := range pods.Items {
		log.Info(p.Status.PodIP)
	}

	configMap := &corev1.ConfigMap{}
	err = s.Get(ctx, r.NamespacedName, configMap)

	return reconcile.Result{}, nil
}
