package utils

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetJobNameAndItemNameFromObject(object client.Object) (string, string) {
	annotations := object.GetAnnotations()
	return annotations[CreateByJob], annotations[CreateByJobItem]
}

func CalJobItemSubName(jobName, itemName, baseName string) string {
	return fmt.Sprintf("%s-%s-%s", jobName, itemName, baseName)
}
