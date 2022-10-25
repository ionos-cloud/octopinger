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
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
}

type daemonReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile ...
func (d *daemonReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	o := &v1alpha1.Octopinger{}

	err := d.Get(ctx, r.NamespacedName, o)
	if err != nil {
		return reconcile.Result{}, err
	}

	nodeList := &corev1.NodeList{}
	err = d.List(ctx, nodeList)
	if err != nil {
		return reconcile.Result{}, err
	}

	configMap := &corev1.ConfigMap{}
	err = d.Get(ctx, types.NamespacedName{Name: o.Name + "-config", Namespace: o.Namespace}, configMap, &client.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	cfg := o.Spec.Probes.ConfigMap()
	cfg["nodes"] = ""

	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name + "-config",
			Namespace: o.Namespace,
		},
		Data: cfg,
	}

	err = d.Create(ctx, configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	}

	err = controllerutil.SetControllerReference(o, configMap, d.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	items := []corev1.KeyToPath{}
	for k := range cfg {
		items = append(items, corev1.KeyToPath{Key: k, Path: k})
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
						"daemonset":  o.Name + "-daemonset",
						"octopinger": o.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "octopinger-container",
							ImagePullPolicy: corev1.PullAlways,
							Image:           o.Spec.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-vol",
									MountPath: "/etc/config",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"NET_RAW",
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config-vol",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: o.Name + "-config",
									},
									Items: items,
								},
							},
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
	err = d.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	if errors.IsNotFound(err) {
		log.Info("creating daemonset")

		err = d.Create(ctx, deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		log.Info("updating daemonset")

		found.Spec = deploy.Spec
		err = d.Update(ctx, found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
