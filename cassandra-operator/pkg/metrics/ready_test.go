package metrics

import (
	"testing"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test(t *testing.T) {
	c, err := cluster.New(
		&v1alpha1.Cassandra{
			Spec: v1alpha1.CassandraSpec{
				UseEmptyDir: ptr.Bool(true),
				Racks: []v1alpha1.Rack{
					{
						Name:     "rack1",
						Zone:     "a",
						Replicas: 1,
					},
				},
				Pod: v1alpha1.Pod{
					Memory: resource.MustParse("1Gi"),
				},
			},
		},
	)
	require.NoError(t, err)
	nt := NewNodetool(c)
	ready, err := nt.IsLocalNodeReady()
	require.NoError(t, err)
	assert.True(t, ready)
}
