package controllers

import (
	"context"
	"reflect"

	v1alpha1 "github.com/ionos-cloud/octopinger/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
			scheme: mgr.GetScheme(),
		})
}

type daemonReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (d *daemonReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	o := &v1alpha1.Octopinger{}

	err := d.client.Get(ctx, r.NamespacedName, o)
	if err != nil {
		return reconcile.Result{}, err
	}

	deploy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name + "-daemonset",
			Namespace: o.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"daemonset":  o.Name + "-daemonset",
					"octopinger": o.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"daemonset":  o.Name + "-daemonset-",
						"octopinger": o.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "octopinger-container",
							Image: o.Spec.Image,
						},
					},
				},
			},
		},
	}

	log = log.WithValues("octopinger", deploy.ObjectMeta.Name)

	err = controllerutil.SetControllerReference(o, deploy, d.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	found := &appsv1.DaemonSet{}
	err = d.client.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	if errors.IsNotFound(err) {
		log.Info("creating daemonset")

		err = d.client.Create(ctx, deploy)
		if err != nil {
			log.Error(err, "could not create daemonset")

			return reconcile.Result{}, err
		}
	}

	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		log.Info("updating daemonset")

		found.Spec = deploy.Spec
		err = d.client.Update(ctx, found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
