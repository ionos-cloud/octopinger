package controller

import (
	"context"
	"time"

	"github.com/ionos-cloud/octopinger/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("DaemonSet controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		OctopingerName      = "test-octopinger"
		OctopingerNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Octopinger Status", func() {
		It("Should incre", func() {
			By("By creating a new Octopinger")
			ctx := context.Background()
			cronJob := &v1alpha1.Octopinger{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.GroupVersion.String(),
					Kind:       v1alpha1.CRDResourceKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      OctopingerName,
					Namespace: OctopingerNamespace,
				},
				Spec: v1alpha1.OctopingerSpec{
					Template: v1alpha1.Template{
						Image: "ghcr.io/ionos-cloud/octopinger/octopinger:v0.1.7-beta.3",
					},
					Config: v1alpha1.Config{
						ICMP: v1alpha1.ICMP{
							Enable: true,
						},
						DNS: v1alpha1.DNS{
							Enable: false,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, cronJob)).Should(Succeed())

			octopingerLookupKey := types.NamespacedName{Name: OctopingerName, Namespace: OctopingerNamespace}
			createdOctopinger := &v1alpha1.Octopinger{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, octopingerLookupKey, createdOctopinger)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdOctopinger.Name).Should(Equal(OctopingerName))
		})
	})

})
