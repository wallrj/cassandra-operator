package validation_test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kr/pretty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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

func set(d interface{}, path string, value interface{}) {
	segments := strings.Split(path, ".")
	v := reflect.ValueOf(d)
	for _, s := range segments {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		v = index(v, s)
	}
	v.Set(reflect.ValueOf(value))
}

func index(v reflect.Value, idx string) reflect.Value {
	if i, err := strconv.Atoi(idx); err == nil {
		return v.Index(i)
	}
	return v.FieldByName(idx)
}

type mutator func(*v1alpha1.Cassandra)

func mutate(path string, value interface{}) mutator {
	return func(c *v1alpha1.Cassandra) {
		set(c, path, value)
	}
}

var _ = Describe("validation functions", func() {
	Context("ValidateCassandra", func() {
		var cass *v1alpha1.Cassandra
		BeforeEach(func() {
			cass = &v1alpha1.Cassandra{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster1",
					Namespace: "ns1",
				},
				Spec: v1alpha1.CassandraSpec{
					Racks: []v1alpha1.Rack{
						{
							Name:         "rack1",
							Zone:         "zone1",
							StorageClass: "fast",
							Replicas:     1,
						},
					},
				},
			}
		})
		DescribeTable(
			"success cases",
			func(mutations ...mutator) {
				for _, mutate := range mutations {
					mutate(cass)
				}
				By(pretty.Sprintf("validating %# v", *cass))
				err := validation.ValidateCassandra(cass).ToAggregate()
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("complete cluster"),
		)

		DescribeTable(
			"failure cases",
			func(mutations ...mutator) {
				for _, mutate := range mutations {
					mutate(cass)
				}
				By(pretty.Sprintf("validating %# v", *cass))
				err := validation.ValidateCassandra(cass).ToAggregate()
				fmt.Fprintf(GinkgoWriter, "INFO: Error message was: %s", err)
				Expect(err).To(HaveOccurred())
			},
			// ObjectMeta
			Entry("missing name", mutate("ObjectMeta.Name", "")),
			Entry("missing namespace", mutate("ObjectMeta.Namespace", "")),

			// Spec.Racks
			Entry("missing racks", mutate("Spec.Racks", []v1alpha1.Rack{})),
			Entry("rack with 0 replicas", mutate("Spec.Racks.0.Replicas", int32(0))),
			Entry(
				"rack with empty storageClass",
				mutate("Spec.UseEmptyDir", ptr.Bool(false)),
				mutate("Spec.Racks.0.StorageClass", ""),
			),
			Entry(
				"rack with empty zone",
				mutate("Spec.UseEmptyDir", ptr.Bool(false)),
				mutate("Spec.Racks.0.Zone", ""),
			),
		)
	})
})
