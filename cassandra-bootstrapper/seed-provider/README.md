# Cassandra SeedProvider for use in Kubernetes

This project implements a rack-aware SeedProvider which is used with Cassandra clusters which have been provisioned by
the [cassandra-operator](../../cassandra-operator).

## Configuration

In order to use this seed provider, the following parameters must be supplied in the `seed_provider` config block in
`cassandra.yaml`. All properties are mandatory and must be non-empty.

Property name  | Type                    | Meaning
---------------|-------------------------|-----------------------------------------------------------------
`namespace`    | string                  | Kubernetes namespace in which the Cassandra cluster is deployed.
`clusterName`  | string                  | Name given to the Cassandra cluster.
`racks`        | comma-separated strings | List of the identifiers of available racks 

## Seed assignment

The assignment algorithm varies depending on the number of available racks:

### For one rack

It works by designating a certain number of the provisioned nodes as seeds.

```
min(floor(clusterSize / 2), 3)
```

In other words, the number of seeds in a cluster will be half of the cluster size, up to a maximum of 3 seeds.

### For multiple racks

One node in each rack will be designated as a seed.