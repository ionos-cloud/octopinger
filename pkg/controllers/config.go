package controllers

import (
	"context"
	"strings"

	"github.com/ionos-cloud/octopinger/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NewConfigReconciler ...
func NewConfigReconciler(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Owns(&corev1.Pod{}).
		WithEventFilter(OcotopingerManaged()).
		Complete(&configReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
}

type configReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func OcotopingerManaged() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		refs := object.GetOwnerReferences()

		for _, r := range refs {
			if r.Kind == "Octopinger" {
				return false
			}
		}

		return true
	})
}

// Reconcile ...
func (s *configReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	ds := &appsv1.DaemonSet{}
	err := s.Get(ctx, r.NamespacedName, ds)
	if err != nil {
		return reconcile.Result{}, nil
	}

	o := &v1alpha1.Octopinger{}
	err = s.Get(ctx, r.NamespacedName, o)
	if err != nil {
		return reconcile.Result{}, err
	}

	pods := &corev1.PodList{}
	err = s.List(ctx, pods, client.InNamespace(r.Namespace), client.MatchingLabels(ds.Spec.Template.Labels))
	if err != nil {
		return ctrl.Result{}, err
	}

	ips := make([]string, 0)
	for _, p := range pods.Items {
		ips = append(ips, p.Status.PodIP)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name + "-config",
			Namespace: r.Namespace,
		},
		Data: map[string]string{
			"nodes": strings.Join(ips, ","),
		},
	}

	err = s.Update(ctx, configMap)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
