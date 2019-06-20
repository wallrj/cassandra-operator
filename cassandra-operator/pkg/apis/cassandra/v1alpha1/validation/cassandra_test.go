package validation_test

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/validation"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

func TestCassandra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Cluster Suite", test.CreateParallelReporters("cluster"))
}

var _ = Describe("validation functions", func() {
	Context("ValidateCassandra", func() {
		var (
			cass *v1alpha1.Cassandra
			err  error
		)

		BeforeEach(func() {
			err = nil
			cass = &v1alpha1.Cassandra{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster1",
					Namespace: "ns1",
				},
				Spec: v1alpha1.CassandraSpec{
					UseEmptyDir: ptr.Bool(false),
					Racks: []v1alpha1.Rack{
						{
							Name:         "rack1",
							Zone:         "zone1",
							StorageClass: "fast",
							Replicas:     1,
						},
					},
					Pod: v1alpha1.Pod{
						CPU:         resource.MustParse("0"),
						Memory:      resource.MustParse("1Gi"),
						StorageSize: resource.MustParse("100Gi"),
					},
				},
			}
		})

		Context("success cases", func() {
			AfterEach(func() {
				By(pretty.Sprintf("validating %# v", *cass))
				err = validation.ValidateCassandra(cass).ToAggregate()
				Expect(err).ToNot(HaveOccurred())
			})
			It("succeeds with a fully populated Cassandra object", func() {})
		})

		Context("failure cases", func() {
			AfterEach(func() {
				By(pretty.Sprintf("validating %# v", *cass))
				err = validation.ValidateCassandra(cass).ToAggregate()
				fmt.Fprintf(GinkgoWriter, "INFO: Error message was: %s", err)
				Expect(err).To(HaveOccurred())
			})

			Context("ObjectMeta", func() {
				It("fails if name is missing", func() {
					cass.Name = ""
				})
				It("fails if namespace is missing", func() {
					cass.Namespace = ""
				})
			})

			Context("Spec", func() {
				Context("Racks", func() {
					It("fails if racks is empty", func() {
						cass.Spec.Racks = nil
					})
					It("fails if Rack.Replicas is < 1", func() {
						cass.Spec.Racks[0].Replicas = 0
					})
					Context("UseEmptyDir=false", func() {
						BeforeEach(func() {
							cass.Spec.UseEmptyDir = ptr.Bool(false)
						})
						It("fails if Rack.StorageClass is empty", func() {
							cass.Spec.Racks[0].StorageClass = ""
						})
						It("fails if Rack.Zone is empty", func() {
							cass.Spec.Racks[0].Zone = ""
						})
					})
				})

				Context("Pod", func() {
					It("fails if Memory is zero", func() {
						cass.Spec.Pod.Memory.Set(0)
					})
					Context("UseEmptyDir=false", func() {
						BeforeEach(func() {
							cass.Spec.UseEmptyDir = ptr.Bool(false)
						})
						It("fails if StorageSize is zero", func() {
							cass.Spec.Pod.StorageSize.Set(0)
						})
					})
					Context("UseEmptyDir=true", func() {
						BeforeEach(func() {
							cass.Spec.UseEmptyDir = ptr.Bool(true)
						})
						It("fails if StorageSize is not zero", func() {
							cass.Spec.Pod.StorageSize.Set(1)
						})
					})
				})
			})
		})
	})
})
