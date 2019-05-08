package operations

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/dispatcher"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestOperations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Operations Suite", test.CreateParallelReporters("operations"))
}

var _ = Describe("operations to execute based on event", func() {
	var (
		receiver      *Receiver
		oldClusterDef *v1alpha1.Cassandra
		newClusterDef *v1alpha1.Cassandra
		timeout       = int32(1)
		sevenDays     = int32(7)
		clusters      = map[string]*cluster.Cluster{}
	)

	BeforeEach(func() {
		oldClusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metav1.ObjectMeta{Name: "mycluster", Namespace: "mynamespace"},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
				Snapshot: &v1alpha1.Snapshot{
					Schedule:       "2 * * * *",
					TimeoutSeconds: &timeout,
					RetentionPolicy: &v1alpha1.RetentionPolicy{
						Enabled:               true,
						CleanupSchedule:       "1 * * * *",
						CleanupTimeoutSeconds: &timeout,
						RetentionPeriodDays:   &sevenDays,
					},
				},
			},
		}
		newClusterDef = oldClusterDef.DeepCopy()

		c, _ := cluster.New(newClusterDef)
		clusters[newClusterDef.QualifiedName()] = c
		receiver = NewEventReceiver(clusters, &cluster.Accessor{}, &metrics.PrometheusMetrics{}, &stubEventRecorder{})
	})

	Context("when a cluster is added", func() {
		It("should return an add cluster operation", func() {
			// given
			newClusterDef.Spec.Snapshot = nil

			// when
			operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCluster, Data: newClusterDef})

			//then
			Expect(operations).To(HaveLen(1))
			Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddClusterOperation{})))
		})

		Context("a snapshot spec exists", func() {

			It("should return add cluster and add snapshot operations", func() {
				// given
				newClusterDef.Spec.Snapshot.RetentionPolicy = nil

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(2))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
			})

			It("should return add cluster and add snapshot operations when the retention policy is disabled", func() {
				// given
				newClusterDef.Spec.Snapshot.RetentionPolicy.Enabled = false

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(2))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
			})

			It("should return add cluster, add snapshot and add cleanup snapshot operations when a snapshot retention spec exists", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(3))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
				Expect(reflect.TypeOf(operations[2])).To(Equal(reflect.TypeOf(&AddSnapshotCleanupOperation{})))
			})
		})

	})

	Context("when a cluster is deleted", func() {
		It("should return a delete cluster operation", func() {
			// given
			newClusterDef.Spec.Snapshot = nil

			// when
			operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCluster, Data: newClusterDef})

			//then
			Expect(operations).To(HaveLen(1))
			Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&DeleteClusterOperation{})))
		})

		Context("a snapshot spec exists", func() {
			It("should return a delete cluster and delete snapshot operations", func() {
				// given
				newClusterDef.Spec.Snapshot.RetentionPolicy = nil

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(2))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&DeleteClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotOperation{})))
			})
			It("should return a delete cluster, delete snapshot and delete snapshot cleanup operations when a retention policy is defined", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(3))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&DeleteClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotOperation{})))
				Expect(reflect.TypeOf(operations[2])).To(Equal(reflect.TypeOf(&DeleteSnapshotCleanupOperation{})))
			})
			It("should return a delete cluster and delete snapshot operations when a retention policy is disabled", func() {
				// given
				newClusterDef.Spec.Snapshot.RetentionPolicy.Enabled = false

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCluster, Data: newClusterDef})

				//then
				Expect(operations).To(HaveLen(2))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&DeleteClusterOperation{})))
				Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotOperation{})))
			})

		})
	})

	Context("when a cluster spec is updated", func() {
		Context("no record was found for the cluster", func() {
			BeforeEach(func() {
				delete(clusters, newClusterDef.QualifiedName())
			})

			It("should return an add cluster operation", func() {
				// given
				newClusterDef.Spec.Snapshot = nil

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddClusterOperation{})))
			})
		})

		Context("a record was found for the cluster", func() {
			It("should return an update cluster operation", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
			})

			Context("a snapshot or snapshot cleanup spec is removed", func() {
				It("should return update cluster, delete snapshot and delete snapshot cleanup when snapshot spec is removed", func() {
					// given
					newClusterDef.Spec.Snapshot = nil

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(3))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotOperation{})))
					Expect(reflect.TypeOf(operations[2])).To(Equal(reflect.TypeOf(&DeleteSnapshotCleanupOperation{})))
				})
				It("should return update cluster and delete snapshot cleanup when snapshot retention policy is removed", func() {
					// given
					newClusterDef.Spec.Snapshot.RetentionPolicy = nil

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotCleanupOperation{})))
				})
				It("should return update cluster and delete snapshot cleanup when snapshot retention policy is disabled", func() {
					// given
					newClusterDef.Spec.Snapshot.RetentionPolicy.Enabled = false

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&DeleteSnapshotCleanupOperation{})))
				})
			})

			Context("a snapshot or snapshot cleanup is added", func() {
				It("should return update cluster, add snapshot and add snapshot cleanup when snapshot with retention policy is added", func() {
					// given
					oldClusterDef.Spec.Snapshot = nil

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(3))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
					Expect(reflect.TypeOf(operations[2])).To(Equal(reflect.TypeOf(&AddSnapshotCleanupOperation{})))
				})
				It("should return update cluster and add snapshot when snapshot without retention policy is added", func() {
					// given
					oldClusterDef.Spec.Snapshot = nil
					newClusterDef.Spec.Snapshot.RetentionPolicy = nil

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
				})
				It("should return update cluster and add snapshot when snapshot with retention policy disabled", func() {
					// given
					oldClusterDef.Spec.Snapshot = nil
					newClusterDef.Spec.Snapshot.RetentionPolicy.Enabled = false

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&AddSnapshotOperation{})))
				})
			})

			Context("a snapshot is updated", func() {
				It("should return update cluster and update snapshot", func() {
					// given
					newClusterDef.Spec.Snapshot.Schedule = "1 13 4 * *"

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&UpdateSnapshotOperation{})))
				})
			})

			Context("a snapshot retention policy is updated", func() {
				It("should return update cluster and update snapshot cleanup", func() {
					// given
					newClusterDef.Spec.Snapshot.RetentionPolicy.CleanupSchedule = "1 13 4 * *"

					// when
					operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCluster, Data: ClusterUpdate{OldCluster: oldClusterDef, NewCluster: newClusterDef}})

					// then
					Expect(operations).To(HaveLen(2))
					Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateClusterOperation{})))
					Expect(reflect.TypeOf(operations[1])).To(Equal(reflect.TypeOf(&UpdateSnapshotCleanupOperation{})))
				})
			})
		})

		Context("when gathering metrics is requested", func() {
			It("should return a gather metrics operation", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: GatherMetrics, Data: clusters[newClusterDef.QualifiedName()]})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&GatherMetricsOperation{})))
			})
		})

		Context("when a custom configmap is updated", func() {
			var configMap *corev1.ConfigMap

			BeforeEach(func() {
				configMap = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "mycluster-config", Namespace: "mynamespace"}}
			})

			It("should return an update custom config operation", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&UpdateCustomConfigOperation{})))
			})
			It("should return no operations when no associated cluster is found", func() {
				// given
				delete(clusters, newClusterDef.QualifiedName())

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: UpdateCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(0))
			})
		})

		Context("when a custom configmap is added", func() {
			var configMap *corev1.ConfigMap

			BeforeEach(func() {
				configMap = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "mycluster-config", Namespace: "mynamespace"}}
			})

			It("should return an add custom config operation", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&AddCustomConfigOperation{})))
			})
			It("should return no operations when no associated cluster is found", func() {
				// given
				delete(clusters, newClusterDef.QualifiedName())

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: AddCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(0))
			})
		})

		Context("when a custom configmap is deleted", func() {
			var configMap *corev1.ConfigMap

			BeforeEach(func() {
				configMap = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "mycluster-config", Namespace: "mynamespace"}}
			})

			It("should return a delete custom config operation", func() {
				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(1))
				Expect(reflect.TypeOf(operations[0])).To(Equal(reflect.TypeOf(&DeleteCustomConfigOperation{})))
			})
			It("should return no operations when no associated cluster is found", func() {
				// given
				delete(clusters, newClusterDef.QualifiedName())

				// when
				operations := receiver.operationsToExecute(&dispatcher.Event{Kind: DeleteCustomConfig, Data: configMap})

				// then
				Expect(operations).To(HaveLen(0))
			})
		})

	})
})

type stubEventRecorder struct{}

func (r *stubEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {}
func (r *stubEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *stubEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *stubEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}
