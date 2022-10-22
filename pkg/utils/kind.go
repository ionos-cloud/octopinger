package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ErrHoldFinalizer indicates that the finalizer should be held.
var (
	ErrHoldFinalizer  = errors.New("hold finalizer")
	ErrSAHasNoSecrets = errors.New("service account has no secrets")
)

// HasFinalizer returns true if the given finalizer is present on the object.
func HasFinalizer(obj metav1.Object, finalizer string) bool {
	for _, item := range obj.GetFinalizers() {
		if item == finalizer {
			return true
		}
	}
	return false
}

// EnsureFinalizer appends the given finalizer. If it did not exist before the
// object is updated using the client.
func EnsureFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizer {
			return nil
		}
	}

	finalizers = append(finalizers, finalizer)
	obj.SetFinalizers(finalizers)
	return c.Update(ctx, obj)
}

// EnsureNoFinalizer removes the given finalizer. If it existed the object is
// updated using the client.
func EnsureNoFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	finalizers := obj.GetFinalizers()
	var (
		filtered []string
		found    bool
	)
	for _, f := range finalizers {
		if f == finalizer {
			found = true
		} else {
			filtered = append(filtered, f)
		}
	}
	if !found {
		return nil
	}
	obj.SetFinalizers(filtered)
	return c.Update(ctx, obj)
}

// WithFinalizer ensures that finalizer is on obj until obj is deleted.
// Delete is called when obj is deleted. The finalizer is removed if
// delete return no error.
func WithFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string, reconcile, delete func() (ctrl.Result, error)) (ctrl.Result, error) {
	if obj.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(obj, finalizer) {
			return reconcile()
		}

		controllerutil.AddFinalizer(obj, finalizer)
		if err := c.Update(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}

		// Don't call reconcile immediately, to follow the best practice of only
		// one object update per reconciliation.
		return ctrl.Result{Requeue: true}, nil
	}

	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		return ctrl.Result{}, nil
	}

	if result, err := delete(); err != nil {
		if errors.Is(err, ErrHoldFinalizer) {
			return result, nil
		}
		return result, err
	}

	// TODO: Also try to follow the best practice of only one update for the deletion branch.
	// For now, we're fine, as we don't do updates during deletion, however, this may strike us.
	controllerutil.RemoveFinalizer(obj, finalizer)
	return ctrl.Result{}, c.Update(ctx, obj)
}

// SetLabel is a helper function that sets the given label and creates the
// labels map if it does not exist yet.
func SetLabel(obj metav1.Object, key, value string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[key] = value
	obj.SetLabels(labels)
}

// SetAnnotation is a helper function that sets the given annotation and creates the
// annotation map if it does not exist yet.
func SetAnnotation(obj metav1.Object, key, value string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[key] = value
	obj.SetAnnotations(annotations)
}

// RESTConfigFromServiceAccountSecret constructs a rest config using the secret
// and host provided.
func RESTConfigFromServiceAccountSecret(secret *corev1.Secret, kubernetesAddress string) (*rest.Config, error) {
	const (
		tokenKey  = "token"
		rootCAKey = "ca.crt"
	)

	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("secret did not contain token data at %s", tokenKey)
	}

	rootCA, ok := secret.Data[rootCAKey]
	if !ok {
		return nil, fmt.Errorf("secret did not contain root CA data at %s", rootCAKey)
	}

	tlsClientConfig := rest.TLSClientConfig{
		CertData: rootCA,
	}

	return &rest.Config{
		Host:            kubernetesAddress,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
	}, nil
}

// ClientConfigFromServiceAccountKey fetches the service account secret and
// calls ClientConfigFromServiceAccountSecret.
func ClientConfigFromServiceAccountKey(
	ctx context.Context,
	log logr.Logger,
	k8sClient client.Client,
	accountKey client.ObjectKey,
	externalKubernetesAddress string,
) (*clientcmdapi.Config, error) {
	account := &corev1.ServiceAccount{}
	if err := k8sClient.Get(ctx, accountKey, account); err != nil {
		return nil, fmt.Errorf("could not find service account %s for fragment: %w", accountKey, err)
	}
	if len(account.Secrets) == 0 {
		return nil, fmt.Errorf("%s/%s: %w", account.Namespace, account.Name, ErrSAHasNoSecrets)
	}
	secret := &corev1.Secret{}
	namespace := account.Secrets[0].Namespace
	if namespace == "" {
		namespace = account.Namespace
	}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: account.Secrets[0].Name}, secret); err != nil {
		return nil, err
	}

	return ClientConfigFromServiceAccountSecret(secret, externalKubernetesAddress)
}

// ClientConfigFromServiceAccountSecret uses the secret and address to construct
// a clientconfig that can be used to create a KUBECONFIG.
func ClientConfigFromServiceAccountSecret(secret *corev1.Secret, kubernetesAddress string) (*clientcmdapi.Config, error) {
	const (
		tokenKey     = "token"
		rootCAKey    = "ca.crt"
		namespaceKey = "namespace"
	)

	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("secret did not contain token data at %s", tokenKey)
	}

	rootCA, ok := secret.Data[rootCAKey]
	if !ok {
		return nil, fmt.Errorf("secret did not contain root CA data at %s", rootCAKey)
	}

	namespace, ok := secret.Data[namespaceKey]
	if !ok {
		return nil, fmt.Errorf("secret did not contain namespace at %s", namespaceKey)
	}

	return &clientcmdapi.Config{
		Preferences: clientcmdapi.Preferences{},
		Clusters: map[string]*clientcmdapi.Cluster{
			"default": {
				Server:                   kubernetesAddress,
				CertificateAuthorityData: rootCA,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"default": {
				Token: string(token),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"default": {
				Cluster:   "default",
				AuthInfo:  "default",
				Namespace: string(namespace),
			},
		},
		CurrentContext: "default",
	}, nil
}

func ClientFromKubeconfig(kubeconfig []byte, opts client.Options) (client.Client, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, opts)
}

// MergeTaints returns a new list of taints after merging all the new taints. If
// a taint with the same key already existed it will be overwritten, otherwise
// it will be appended.
func MergeTaints(oldTaints []corev1.Taint, newTaints ...corev1.Taint) []corev1.Taint {
	res := make([]corev1.Taint, len(oldTaints))
	copy(res, oldTaints)

	for _, newTaint := range newTaints {
		var found bool
		for i, oldTaint := range oldTaints {
			if oldTaint.Key == newTaint.Key {
				res[i] = newTaint
				found = true
			}
		}
		if !found {
			res = append(res, newTaint)
		}
	}
	return res
}

// ObjectKeyToObjectMeta returns a new metav1.ObjectMeta for the object key.
func ObjectKeyToObjectMeta(key client.ObjectKey) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: key.Namespace,
		Name:      key.Name,
	}
}

// MetaObjectToObjectKey returns a new client.ObjectKey for the meta object.
func MetaObjectToObjectKey(meta metav1.Object) client.ObjectKey {
	return client.ObjectKey{
		Namespace: meta.GetNamespace(),
		Name:      meta.GetName(),
	}
}

// FindContainer finds a container by name. Second return value will be false, if
// no matching container was found.
func FindContainer(containers []corev1.Container, name string) (*corev1.Container, bool) {
	for i := 0; i < len(containers); i++ {
		container := &containers[i]
		if container.Name == name {
			return container, true
		}
	}
	return nil, false
}

// CreateOrUpdateContainer either creates or finds a container with the given name. The
// desired state of that container should be reconciled inside updateFn. The
// updated container slice is returned.
func CreateOrUpdateContainer(containers []corev1.Container, name string, updateFn func(c *corev1.Container) error) ([]corev1.Container, error) {
	res := make([]corev1.Container, len(containers))
	copy(res, containers)

	c, found := FindContainer(res, name)
	if !found {
		res = append(res, corev1.Container{
			Name: name,
		})
		c = &res[len(res)-1]
	}
	if err := updateFn(c); err != nil {
		return nil, err
	}
	return res, nil
}

// MergeEnvVars returns a new slice of EnvVar after merging all the new envvars. If
// a envvar with the same name already existed it will be overwritten, otherwise
// it will be appended.
func MergeEnvVars(oldEnvs []corev1.EnvVar, newEnvs ...corev1.EnvVar) []corev1.EnvVar {
	res := make([]corev1.EnvVar, len(oldEnvs))
	copy(res, oldEnvs)

	for _, newEnv := range newEnvs {
		var found bool
		for i, oldEnv := range oldEnvs {
			if oldEnv.Name == newEnv.Name {
				res[i] = newEnv
				found = true
			}
		}
		if !found {
			res = append(res, newEnv)
		}
	}
	return res
}

// SetInheritedLabels sets the labels from the parent if they exist, overwriting
// labels that existed before.
func SetInheritedLabels(obj metav1.Object, parent metav1.Object, keys ...string) {
	parentLabels := parent.GetLabels()
	for _, key := range keys {
		if value, ok := parentLabels[key]; ok {
			SetLabel(obj, key, value)
		}
	}
}

// ParseNamespacedNamed turns "namespace/resource" into a NamespacedName.
func ParseNamespacedNamed(nn string) (types.NamespacedName, error) {
	parts := strings.Split(nn, string(types.Separator))
	if len(parts) != 2 {
		return types.NamespacedName{}, fmt.Errorf("namespaced name must have two parts separated by %s", string(types.Separator))
	}
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

// LabelSeparator to use instead of types.Separator.
const LabelSeparator = "_"

// NamespacedNameToLabel encodes a namespacedname into a label-compliant string
// ("namespace_resource").
func NamespacedNameToLabel(nn types.NamespacedName) string {
	return strings.ReplaceAll(nn.String(), string(types.Separator), LabelSeparator)
}

// LabelToNamespacedName reads from a label-encoded name ("namespace_resource")
// and parses the namespaced name.
func LabelToNamespacedName(l string) (types.NamespacedName, error) {
	parts := strings.Split(l, LabelSeparator)
	if len(parts) != 2 {
		return types.NamespacedName{}, fmt.Errorf("namespaced name must have two parts separated by %s", LabelSeparator)
	}
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

// Deleter can delete resources from kubernetes.
type Deleter interface {
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
}

// EnsureResourceDeletion deletes all resources and returns an error if not all could be deleted.
func EnsureResourceDeletion(ctx context.Context, clusterClient client.Client, resourcesToDelete ...client.Object) error {
	var gone int
	for _, resourceToDelete := range resourcesToDelete {
		if err := clusterClient.Delete(ctx, resourceToDelete); client.IgnoreNotFound(err) != nil {
			return err
		}
		gone++
	}

	if gone == len(resourcesToDelete) {
		return nil
	}
	return fmt.Errorf("not all resources were deleted")
}

// GetterDeleter can get and delete resources from kubernetes.
type GetterDeleter interface {
	Deleter
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
}

// EnsureResourceGone calls Delete() only of the object still exists, so it is
// safe to use during reconciliation, as it will not cause additional load on
// the API Server.
func EnsureResourceGone(ctx context.Context, k8sClient GetterDeleter, obj client.Object) error {
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !obj.GetDeletionTimestamp().IsZero() {
		return nil
	}
	ctrl.LoggerFrom(ctx).Info("Deleting not needed resource.", "resourceName", obj.GetName(), "resourceNamespace", obj.GetNamespace(), "resourceKind", GetKind(obj))
	return client.IgnoreNotFound(k8sClient.Delete(ctx, obj))
}

// UpdateSecretData writes the provided fields in the Data field of the struct.
// Provided values should either be []byte or string.
func UpdateSecretData(secret *corev1.Secret, data map[string]interface{}) {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	for key, value := range data {
		secret.Data[key] = []byte(fmt.Sprintf("%s", value))
	}
}

// IsOwnedBy checks if the  object has a owner ref set to the given owner.
func IsOwnedBy(obj metav1.Object, owner metav1.Object) bool {
	refs := obj.GetOwnerReferences()
	for i := range refs {
		if refs[i].UID == owner.GetUID() {
			return true
		}
	}
	return false
}

// IsControlledBy checks if the  object has a controller ref set to the given owner.
func IsControlledBy(obj metav1.Object, owner metav1.Object) bool {
	refs := obj.GetOwnerReferences()
	for i := range refs {
		if refs[i].UID != owner.GetUID() {
			continue
		}
		return pointer.BoolDeref(refs[i].Controller, false)
	}
	return false
}

// LabelSelectorRequirement is an unchecked representation of labels.Requirements.
type LabelSelectorRequirement struct {
	Key  string
	Op   selection.Operator
	Vals []string
}

// NewRequirement creates a new requirement to be used for BuildLabelSelector.
func NewRequirement(key string, op selection.Operator, vals []string) LabelSelectorRequirement {
	return LabelSelectorRequirement{
		Key:  key,
		Op:   op,
		Vals: vals,
	}
}

// BuildLabelSelector builds a label selector or returns an error if any of the requirements are invalid.
func BuildLabelSelector(requirements ...LabelSelectorRequirement) (labels.Selector, error) {
	selector := labels.NewSelector()
	for _, rawReq := range requirements {
		req, err := labels.NewRequirement(rawReq.Key, rawReq.Op, rawReq.Vals)
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*req)
	}
	return selector, nil
}

// RemoveOwnerReference removes the given owner reference from the given
// object.
func RemoveOwnerReference(ctx context.Context, c client.Client, obj client.Object, owner metav1.Object) error {
	references := obj.GetOwnerReferences()
	updatedRefs := []metav1.OwnerReference{}
	for _, r := range references {
		if owner.GetUID() != r.UID {
			updatedRefs = append(updatedRefs, r)
		}
	}
	obj.SetOwnerReferences(updatedRefs)
	return c.Update(ctx, obj)
}

// NodeAffinityFromSelector converts a simple nodeSelector to a nodeAffinity,
// where the terms are ANDed.
func NodeAffinityFromSelector(nodeSelector map[string]string) *corev1.NodeAffinity {
	if len(nodeSelector) == 0 {
		return nil
	}
	keys := make([]string, 0, len(nodeSelector))
	for k := range nodeSelector {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	requirements := []corev1.NodeSelectorRequirement{}
	for _, k := range keys {
		requirements = append(requirements, corev1.NodeSelectorRequirement{
			Key:      k,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{nodeSelector[k]},
		})
	}
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{MatchExpressions: requirements},
			},
		},
	}
}

// SortTypedLocalObjectReference into a stable order.
func SortTypedLocalObjectReference(refs []corev1.TypedLocalObjectReference) {
	sort.Slice(refs, func(i, j int) bool {
		u1, u2 := refs[i], refs[j]
		return lessNilStrings(u1.APIGroup, u2.APIGroup) ||
			u1.Kind < u2.Kind ||
			u1.Name < u2.Name
	})
}

func lessNilStrings(s1, s2 *string) bool {
	return s1 == nil && s2 != nil ||
		s1 != nil && s2 != nil && *s1 < *s2
}

// UpdateStatusWithResult updates the status of the object, only if the status
// sub-resource has changed, and returns whether it was changed or not.
func UpdateStatusWithResult(ctx context.Context, c client.Client, obj client.Object) (controllerutil.OperationResult, error) {
	none := controllerutil.OperationResultNone
	objStatus, err := getStatus(obj)
	if err != nil {
		return none, err
	}

	currentObj := obj.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), currentObj); err != nil {
		return none, err
	}
	currentObjStatus, err := getStatus(currentObj)
	if err != nil {
		return none, err
	}

	if reflect.DeepEqual(objStatus, currentObjStatus) {
		return none, nil
	}

	if err := c.Status().Update(ctx, obj); err != nil {
		return none, err
	}
	return controllerutil.OperationResultUpdatedStatusOnly, nil
}

// UpdateStatus updates the status of the object, only if the status sub-resource has changed.
func UpdateStatus(ctx context.Context, c client.Client, obj client.Object) error {
	_, err := UpdateStatusWithResult(ctx, c, obj)
	return err
}

func getStatus(obj client.Object) (interface{}, error) {
	// convert object to unstructured data
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	// attempt to extract the status from the resource
	status, _, err := unstructured.NestedFieldNoCopy(data, "status")
	if err != nil {
		return nil, err
	}
	return status, nil
}

// GetKind returns the Kind of the kubernetes object. If the kind is not set
// (e.g. during tests without envtest), it returns the struct name, which should
// usually match the kind.
func GetKind(obj runtime.Object) string {
	if kind := obj.GetObjectKind().GroupVersionKind().Kind; kind != "" {
		return kind
	}
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

// EnqueueRequestFromNameLabel determines the reconcile request name based on
// the value of the specified label. If a namespace is specified, that namespace
// will be used instead of the objects namespace. This is useful if you want to
// reconcile based on changes without using an OwnerReference.
func EnqueueRequestFromNameLabel(label string, namespace ...string) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(object client.Object) []reconcile.Request {
			name, ok := object.GetLabels()[label]
			if !ok || name == "" {
				return nil
			}
			ns := object.GetNamespace()
			if len(namespace) == 1 {
				ns = namespace[0]
			}
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}}}
		},
	)
}

const inClusterNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// GetK8sNamespace returns the namespace the process is running in, when
// actually running in-cluster. Returns fallback otherwise.
func GetK8sNamespace(fallback string) (string, error) {
	_, err := os.Stat(inClusterNamespacePath)
	if os.IsNotExist(err) {
		return fallback, nil
	} else if err != nil {
		return "", err
	}

	// Load the namespace file and return its content
	namespace, err := os.ReadFile(inClusterNamespacePath)
	if err != nil {
		return "", err
	}
	return string(namespace), nil
}

const skipReconciliationLabel = "labels.dbaas.ionos.com/SkipReconciliation"

// SkipReconcile returns true if reconciliation should be temporarily skipped
// for this resource.
func SkipReconcile(obj client.Object) bool {
	return obj.GetLabels()[skipReconciliationLabel] == "true"
}
