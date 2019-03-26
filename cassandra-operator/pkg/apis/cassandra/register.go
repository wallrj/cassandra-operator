package cassandra

const (
	// GroupName is the name of the API group
	// Remember to change this reference in `doc.go` too
	GroupName = "core.sky.uk"

	// Version is the version of the resource
	Version = "v1alpha1"

	// Name is the full name of the resource
	Name = Plural + ".core.sky.uk"

	// Singular is the singular form of the resource name
	Singular = "cassandra"

	// Plural is the plural form of the resource name
	Plural = "cassandras"

	// Kind is the object Kind
	Kind = "Cassandra"
)
