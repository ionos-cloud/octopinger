package controllers

import (
	"context"

	v1alpha1 "github.com/ionos-cloud/octopinger/api/v1alpha1"
	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DaemonReconciler ...
type DaemonReconciler struct {
	client       client.Client
	recoverPanic bool
}

// Recon
func (r *DaemonReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// SetupWithManager ...
func (r *DaemonReconciler) SetupWithManager(mgr manager.Manager) error {
	c, err := controller.New("octopinger", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	err = c.Watch(source.NewKindWithCache(&v1alpha1.Octopinger{}, mgr.GetCache()), &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// c.Watch(&source.Kind{Type: &corev1.Node{}}, )

	objs := []client.Object{
		&appsv1.DaemonSet{},
	}

	for _, obj := range objs {
		err = c.Watch(
			&source.Kind{Type: obj},
			&handler.EnqueueRequestForObject{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DaemonReconciler) predicate(ctx context.Context, log zap.Logger) predicate.Funcs {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		octopinger, ok := object.(*v1alpha1.Octopinger)
		if !ok {
			return false
		}

		if octopinger.Spec.Image == "" {
			return false
		}

		return true
	})
}

// NewDaemonReconciler ...
// func NewDaemonReconciler(client client.Client, recover bool) (*DaemonReconciler, error) {
// 	if client == nil {
// 		return nil, fmt.Errorf("client needs to be set")
// 	}

// 	r := new(DaemonReconciler)
// 	r.client = client
// 	r.recoverPanic = recover

// 	return r, nil
// }

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
