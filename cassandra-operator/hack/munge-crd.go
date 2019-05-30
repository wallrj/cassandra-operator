package main

/*
This script updates a CRD yaml file to work around limitations in kubernetes
1.10. It should not be required if targeting a version equal to or higher than
1.11.
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sigs.k8s.io/yaml"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("usage: go run munge-crd.go path/to/yaml-file.yml")
	}

	path := os.Args[1]

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	var crd apiextensionsv1beta1.CustomResourceDefinition
	err = yaml.Unmarshal(bytes, &crd)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	crd.Spec.Scope = "Namespaced"
	crd.Spec.Version = "v1alpha1"

	// fix spec.validation.openAPIV3Schema.properties[metadata].properties[annotations].additionalProperties: Forbidden: additionalProperties cannot be set to false
	// fix spec.validation.openAPIV3Schema.properties[metadata].properties[labels].additionalProperties: Forbidden: additionalProperties cannot be set to false
	schemaProps := crd.Spec.Validation.OpenAPIV3Schema.Properties["metadata"]
	annotations := schemaProps.Properties["annotations"]
	annotations.AdditionalProperties = nil
	schemaProps.Properties["annotations"] = annotations
	labels := schemaProps.Properties["labels"]
	labels.AdditionalProperties = nil
	schemaProps.Properties["labels"] = labels
	crd.Spec.Validation.OpenAPIV3Schema.Properties["metadata"] = schemaProps

	y, err := yaml.Marshal(crd)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	fmt.Println(string(y))
}
