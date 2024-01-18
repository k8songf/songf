package controller

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getJobNameAndItemNameFromObject(object client.Object) (string, string) {
	annotations := object.GetAnnotations()
	return annotations[CreateByJob], annotations[CreateByJobItem]
}

func calItemSubName(jobName, itemName, baseName string) string {
	return fmt.Sprintf("%s-%s-%s", jobName, itemName, baseName)
}
