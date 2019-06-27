package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	admissionreg "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryyaml "k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("usage: go run munge-webhook.go path/to/yaml-file.yml")
	}

	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	yamlReaderCloser := ioutil.NopCloser(file)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	decoder := apimachineryyaml.NewDocumentDecoder(yamlReaderCloser)

	buffer := bytes.NewBuffer(make([]byte, 0))
	part := make([]byte, 4092)

	var mutator admissionreg.MutatingWebhookConfiguration
	var validator admissionreg.ValidatingWebhookConfiguration
	for {
		count, err := decoder.Read(part)

		if err == io.EOF {
			break
		}
		if err == io.ErrShortBuffer {
			buffer.Write(part[:count])
			continue
		}
		buffer.Write(part[:count])

		res := yaml.MapSlice{}
		yaml.Unmarshal(buffer.Bytes(), &res)

		for _, item := range res {
			key, ok := item.Key.(string)
			if !ok {
				continue
			}
			if key == "kind" {
				kind := item.Value.(string)
				switch kind {
				case "MutatingWebhookConfiguration":
					err = k8syaml.Unmarshal(buffer.Bytes(), &mutator)
					if err != nil {
						log.Fatalf("err: %v\n", err)
					}
				case "ValidatingWebhookConfiguration":
					err = k8syaml.Unmarshal(buffer.Bytes(), &validator)
					if err != nil {
						log.Fatalf("err: %v\n", err)
					}
				default:
					continue
				}
			}
		}
		buffer = bytes.NewBuffer(make([]byte, 0))
	}

	namespaceSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"webhooks.core.sky.uk": "enabled"}}
	for i, _ := range mutator.Webhooks {
		mutator.Webhooks[i].NamespaceSelector = namespaceSelector
	}
	for i, _ := range validator.Webhooks {
		validator.Webhooks[i].NamespaceSelector = namespaceSelector
	}

	y, err := k8syaml.Marshal(mutator)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	fmt.Println(string(y))

	fmt.Println("---")

	y, err = k8syaml.Marshal(validator)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	fmt.Println(string(y))
}
