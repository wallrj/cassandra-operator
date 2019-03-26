package e2e

import (
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const PodCleanupTimeout = 60 * time.Second

func DeleteCassandraPodsInNamespace(namespace string) {
	deletePodsWithLabel(namespace, OperatorLabel)
	gomega.Eventually(cassandraPodsInNamespace(namespace), PodCleanupTimeout, time.Second).Should(gomega.BeZero())
}

func deletePodsWithLabel(namespace, label string) {
	podClient := KubeClientset.CoreV1().Pods(namespace)
	err := podClient.DeleteCollection(&metaV1.DeleteOptions{}, metaV1.ListOptions{LabelSelector: label})
	if err != nil {
		log.Infof("Unable to delete pods labelled %s in namespace %s: %v", label, namespace, err)
	}
}

func cassandraPodsInNamespace(namespace string) func() int {
	return func() int {
		podClient := KubeClientset.CoreV1().Pods(namespace)
		pods, err := podClient.List(metaV1.ListOptions{LabelSelector: OperatorLabel})
		if err != nil {
			return 0
		}

		return len(pods.Items)
	}
}
