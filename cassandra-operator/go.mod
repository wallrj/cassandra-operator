module github.com/sky-uk/cassandra-operator/cassandra-operator

go 1.12

require (
	github.com/PaesslerAG/gval v0.1.1 // indirect
	github.com/PaesslerAG/jsonpath v0.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/gofrs/flock v0.7.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20181024230925-c65c006176ff // indirect
	github.com/google/go-cmp v0.2.0
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/common v0.1.0
	github.com/prometheus/procfs v0.0.0-20190104112138-b1a0a9a36d74 // indirect
	github.com/robfig/cron v1.1.0
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/theckman/go-flock v0.7.0
	golang.org/x/net v0.0.0-20190501004415-9ce7a6920f09 // indirect
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190429190828-d89cdac9e872 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	google.golang.org/appengine v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190502092502-a44ef629a3c9
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	k8s.io/gengo v0.0.0-20181019081622-7338e4bfd691 // indirect
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30 // indirect
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190216013122-f05b8decd79c

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.0-beta.2
