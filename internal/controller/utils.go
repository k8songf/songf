package controller

import "sigs.k8s.io/controller-runtime/pkg/client"

func getJobNameAndItemNameFromObject(object client.Object) (string, string) {
	annotations := object.GetAnnotations()
	return annotations[CreateByJob], annotations[CreateByJobItem]
}
