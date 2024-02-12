package v1alpha1

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

func IsJobItemValid(job *Job) (bool, string) {
	fatherNum := 0
	var itemNames map[string]interface{}

	for _, item := range job.Spec.Items {
		if item.Name == "" {
			return false, "item name can not be nil"
		}

		if len(item.RunAfter) == 0 {
			fatherNum++
		}

		_, ok := itemNames[item.Name]
		if ok {
			return false, fmt.Sprintf("item name %s repeated", item.Name)
		}
		itemNames[item.Name] = nil

		// father item repeated
		if fatherNum > 1 {
			return false, "father num > 0"
		}
	}

	// no father item
	if fatherNum == 0 {
		return false, "job not has father item"
	}

	// parent not found
	for _, item := range job.Spec.Items {
		for _, fatherName := range item.RunAfter {
			if _, ok := itemNames[item.Name]; !ok {
				return false, fmt.Sprintf("item %s parent %s not found", item.Name, fatherName)
			}
		}
	}

	return true, ""
}

func IsJobHasCycle(job *Job) (bool, error) {

	node, _, err := NewGraphFromJob(job)
	if err != nil {
		return false, err
	}

	flag := IsJobHasCycleDfs(node, map[string]bool{})

	return flag, nil
}

func IsJobHasCycleDfs(node *ItemNode, visited map[string]bool) bool {

	if visited[node.Item.Name] {
		return true
	}

	visited[node.Item.Name] = true

	for _, child := range node.Child {
		if IsJobHasCycleDfs(child, visited) {
			return true
		}
	}

	visited[node.Item.Name] = false

	return false
}
