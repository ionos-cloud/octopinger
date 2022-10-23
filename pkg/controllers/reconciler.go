package controllers

import (
	"context"

	v1alpha1 "github.com/ionos-cloud/octopinger/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NewDaemonReconciler ...
func NewDaemonReconciler(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Octopinger{}).
		Watches(
			source.NewKindWithCache(&v1alpha1.Octopinger{}, mgr.GetCache()),
			&handler.EnqueueRequestForObject{}).
		Complete(&daemonReconciler{
			client: mgr.GetClient(),
		})
}

type daemonReconciler struct {
	client client.Client
}

func (d *daemonReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	o := &v1alpha1.Octopinger{}

	err := d.client.Get(ctx, r.NamespacedName, o)
	if err != nil {
		return reconcile.Result{}, err
	}

	// if !kerrors.IsNotFound(err) {
	// 	return reconcile.Result{}, nil
	// }

	// octo := &appsv1.

	return reconcile.Result{}, nil
}
