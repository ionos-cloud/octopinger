package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ionos-cloud/octopinger/api/v1alpha1"
	"github.com/ionos-cloud/octopinger/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NewConfigMapData ..
func NewConfigMapData() ConfigMapData {
	return ConfigMapData{"nodes": "", "config": "{}"}
}

// ConfigMapData ...
type ConfigMapData map[string]string

// SetConfig ...
func (c ConfigMapData) SetConfig(cfg *v1alpha1.Config) error {
	bb, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	c["config"] = string(bb)

	return nil
}

// SetNodes ...
func (c *ConfigMapData) SetNodes() error {
	return nil
}

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
	log.Info("reconcile octopinger")

	octopinger := &v1alpha1.Octopinger{}

	err := d.Get(ctx, r.NamespacedName, octopinger)
	if err != nil && errors.IsNotFound(err) {
		// Request object not found, could have been deleted after reconcile request.
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, err
	}

	// get the latest version of octopinger instance before reconciling
	err = d.Get(ctx, r.NamespacedName, octopinger)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = d.reconcileResources(ctx, octopinger)
	if err != nil {
		// Error reconciling Octopinger sub-resources - requeue the request.
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (d *daemonReconciler) reconcileStatus(ctx context.Context, octopinger *v1alpha1.Octopinger) error {
	phase := v1alpha1.OctopingerPhaseNone

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      octopinger.Name + "-daemonset",
			Namespace: octopinger.Namespace,
		},
	}
	if utils.IsObjectFound(ctx, d, octopinger.Namespace, ds.Name, ds) {
		phase = v1alpha1.OctopingerPhaseCreating

		if ds.Status.CurrentNumberScheduled == ds.Status.DesiredNumberScheduled {
			phase = v1alpha1.OctopingerPhaseRunning
		}
	}

	if octopinger.Status.Phase != phase {
		octopinger.Status.Phase = phase

		return d.Status().Update(ctx, octopinger)
	}

	return nil
}

func (d *daemonReconciler) reconcileDaemonSets(ctx context.Context, octopinger *v1alpha1.Octopinger) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling Octopinger")

	configMap := &corev1.ConfigMap{}
	err := utils.FetchObject(ctx, d, octopinger.Namespace, octopinger.Name+"-config", configMap)
	if err != nil {
		return err
	}

	items := []corev1.KeyToPath{}
	for k := range configMap.Data {
		items = append(items, corev1.KeyToPath{Key: k, Path: k})
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      octopinger.Name + "-daemonset",
			Namespace: octopinger.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"daemonset":  octopinger.Name + "-daemonset",
					"octopinger": octopinger.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"daemonset":  octopinger.Name + "-daemonset",
						"octopinger": octopinger.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "octopinger-container",
							ImagePullPolicy: corev1.PullAlways,
							Image:           octopinger.Spec.Template.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-vol",
									MountPath: "/etc/config",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "status",
									ContainerPort: 8081,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromString("status"),
									},
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
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name: "HOST_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
							},
						},
					},
					Tolerations: octopinger.Spec.Template.Tolerations,
					Volumes: []corev1.Volume{
						{
							Name: "config-vol",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: octopinger.Name + "-config",
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

	err = controllerutil.SetControllerReference(octopinger, ds, d.scheme)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("checking for %s in %s", octopinger.Name+"-daemonset", octopinger.Namespace))

	existingDS := &appsv1.DaemonSet{}
	if utils.IsObjectFound(ctx, d, octopinger.Namespace, octopinger.Name+"-daemonset", existingDS) {
		// this is not DaemonSet is not owned by Octopinger
		if ownerRef := metav1.GetControllerOf(existingDS); ownerRef == nil || ownerRef.Kind != v1alpha1.CRDResourceKind {
			return nil
		}

		if !reflect.DeepEqual(existingDS, ds) {
			existingDS = ds
			return d.Update(ctx, existingDS)
		}

		return nil
	}

	log.Info(fmt.Sprintf("creating %s", octopinger.Name+"-daemonset"))

	return d.Create(ctx, ds)
}

func (d *daemonReconciler) reconcileConfigMaps(ctx context.Context, octopinger *v1alpha1.Octopinger) error {
	log := ctrl.LoggerFrom(ctx)

	log.Info("reconciling config map")

	configMap := &corev1.ConfigMap{}
	if utils.IsObjectFound(ctx, d, octopinger.Namespace, octopinger.Name+"-config", configMap) {
		return nil
	}

	configMapData := NewConfigMapData()
	err := configMapData.SetConfig(&octopinger.Spec.Config)
	if err != nil {
		return err
	}

	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      octopinger.Name + "-config",
			Namespace: octopinger.Namespace,
		},
		Data: configMapData,
	}

	err = controllerutil.SetControllerReference(octopinger, configMap, d.scheme)
	if err != nil {
		return err
	}

	return d.Create(ctx, configMap)
}

func (d *daemonReconciler) reconcileResources(ctx context.Context, octopinger *v1alpha1.Octopinger) error {
	err := d.reconcileStatus(ctx, octopinger)
	if err != nil {
		return err
	}

	err = d.reconcileConfigMaps(ctx, octopinger)
	if err != nil {
		return err
	}

	err = d.reconcileDaemonSets(ctx, octopinger)
	if err != nil {
		return err
	}

	return nil
}
