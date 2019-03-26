#!/usr/bin/env bash
set -e
scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
projectDir="${scriptDir}/../"
buildDir="${projectDir}/build/"

mkdir -p ${buildDir}

function setup_nodes() {
    local start_index=$1
    local end_index=$2
    local zone=$3

    for ((node_index=${start_index}; node_index<=${end_index}; node_index++));
    do
        node="kube-node-${node_index}"
        pv_path="/mnt/pv-zone-${zone}"
        kubectl --context dind label --overwrite node ${node} failure-domain.beta.kubernetes.io/zone=eu-west-1${zone}
        docker exec kube-node-${node_index} mkdir -p /data/vol ${pv_path}/bindmount
        docker exec kube-node-${node_index} mount -o bind /data/vol ${pv_path}/bindmount
    done
}

function create_storage_class() {
    # based on https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/provisioner/deployment/kubernetes/example/default_example_storageclass.yaml
    local zone=$1
    cat <<EOF | kubectl --context dind apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard-zone-${zone}
provisioner: kubernetes.io/no-provisioner
reclaimPolicy: Delete
EOF
}

function create_test_namespace() {
    local namespace=$1
    cat <<EOF | kubectl --context dind apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: ${namespace}
EOF
}

function deploy_local_volume_provisioner_credentials() {
    # based on https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/provisioner/deployment/kubernetes/example/default_example_provisioner_generated.yaml
    cat <<EOF | kubectl --context dind apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: local-storage-admin
  namespace: local-volume-provisioning
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: local-storage-provisioner-pv-binding
  namespace: local-volume-provisioning
subjects:
- kind: ServiceAccount
  name: local-storage-admin
  namespace: local-volume-provisioning
roleRef:
  kind: ClusterRole
  name: system:persistent-volume-provisioner
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: local-storage-provisioner-node-clusterrole
  namespace: local-volume-provisioning
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: local-storage-provisioner-node-binding
  namespace: local-volume-provisioning
subjects:
- kind: ServiceAccount
  name: local-storage-admin
  namespace: local-volume-provisioning
roleRef:
  kind: ClusterRole
  name: local-storage-provisioner-node-clusterrole
  apiGroup: rbac.authorization.k8s.io
EOF
}

function deploy_local_volume_provisioner() {
    # based on https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/provisioner/deployment/kubernetes/example/default_example_provisioner_generated.yaml
    local zone=$1
    cat <<EOF | kubectl --context dind apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-provisioner-config-${zone}
  namespace: local-volume-provisioning
data:
  storageClassMap: |
    standard-zone-${zone}:
       hostDir: /mnt/pv-zone-${zone}
       mountDir: /mnt/pv-zone-${zone}
       blockCleanerCommand:
         - "/scripts/shred.sh"
         - "2"
       volumeMode: Filesystem
       fsType: ext4
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: local-volume-provisioner-${zone}
  namespace: local-volume-provisioning
spec:
  selector:
    matchLabels:
      app: local-volume-provisioner-${zone}
  template:
    metadata:
      labels:
        app: local-volume-provisioner-${zone}
    spec:
      serviceAccountName: local-storage-admin
      containers:
        - image: "quay.io/external_storage/local-volume-provisioner:v2.2.0"
          name: provisioner
          securityContext:
            privileged: true
          env:
          - name: MY_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: MY_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          volumeMounts:
            - mountPath: /etc/provisioner/config
              name: provisioner-config
              readOnly: true
            - mountPath: /mnt/pv-zone-${zone}
              name: pv-zone-${zone}
              mountPropagation: "HostToContainer"
      nodeSelector:
        failure-domain.beta.kubernetes.io/zone: eu-west-1${zone}
      volumes:
        - name: provisioner-config
          configMap:
            name: local-provisioner-config-${zone}
        - name: pv-zone-${zone}
          hostPath:
            path: /mnt/pv-zone-${zone}
EOF
}

function runLocalRegistry {
    # local registry so we can build image locally
    # and tell dind nodes to pull the "local" image from the host
    # based on https://github.com/kubernetes-sigs/kubeadm-dind-cluster/issues/56#issuecomment-387463386
    runningRegistry=$(docker ps --filter=name="dind-registry" --format="{{.Names}}")
    if [[ "$runningRegistry" == "" ]]; then
        echo "Running local registry on port 5000"
        docker run -d --name=dind-registry --rm -p 5000:5000 registry:2
    fi

    echo "Setting up dind nodes to target local registry"
    docker ps -a -q --filter=label=mirantis.kubeadm_dind_cluster | while read container_id;
    do
        docker exec ${container_id} /bin/bash -c \
            "docker rm -fv registry-proxy || true"
        # run registry proxy: forward from localhost:5000 on each node to host:5000
        docker exec ${container_id} /bin/bash -c \
            "docker run --name registry-proxy -d -e LISTEN=':5000' -e TALK=\"\$(/sbin/ip route|awk '/default/ { print \$3 }'):5000\" -p 5000:5000 tecnativa/tcp-proxy"
    done
}

# cluster config
declare -A node_zone_count
node_zone_count=(
    [a]=${ZONE_A_NODES:-"2"}
    [b]=${ZONE_B_NODES:-"2"}
)

read -r -a zones <<< "${!node_zone_count[@]}"
numNodes=0
for numZoneNodes in "${node_zone_count[@]}"
do
    numNodes=$((numNodes + numZoneNodes))
done

DIND_VERSION=${DIND_VERSION:-"1.10"}
echo "Downloading dind artifacts into ${buildDir}"
curl -L "https://github.com/kubernetes-sigs/kubeadm-dind-cluster/releases/download/v0.1.0/dind-cluster-v${DIND_VERSION}.sh" -o "${buildDir}/dind-cluster-v${DIND_VERSION}.sh"
chmod +x ${buildDir}/dind-cluster-v${DIND_VERSION}.sh

# clean previous cluster if present
${buildDir}/dind-cluster-v${DIND_VERSION}.sh clean

# bootstrap cluster
FEATURE_GATES="${FEATURE_GATES:-MountPropagation=true,PersistentLocalVolumes=true}"
KUBELET_FEATURE_GATES="${KUBELET_FEATURE_GATES:-MountPropagation=true,DynamicKubeletConfig=true,PersistentLocalVolumes=true}"

FEATURE_GATES=${FEATURE_GATES} KUBELET_FEATURE_GATES=${KUBELET_FEATURE_GATES} NUM_NODES=${numNodes} MGMT_CIDRS=172.18.0.0/16 POD_NETWORK_CIDR=172.20.0.0/16 ${buildDir}/dind-cluster-v${DIND_VERSION}.sh up

# add kubectl directory to PATH
export PATH="$HOME/.kubeadm-dind-cluster:$PATH"

create_test_namespace local-volume-provisioning
create_test_namespace test-cassandra-operator

deploy_local_volume_provisioner_credentials

start=1
for zone in "${zones[@]}"
do
    end=$((${start} + ${node_zone_count[${zone}]} - 1))
    if [ node_zone_count[${zone}] != 0 ]
    then
        setup_nodes ${start} ${end} ${zone}
    fi
    start=$((${end} + 1))

    create_storage_class ${zone}
    deploy_local_volume_provisioner ${zone}
done

runLocalRegistry
