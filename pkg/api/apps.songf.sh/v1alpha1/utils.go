package v1alpha1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
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
	itemNames := map[string]*Item{}

	for _, item := range job.Spec.Items {
		if item.Name == "" {
			return false, "item name can not be nil"
		}

		flag, msg := IsItemJobResourceValid(item.ItemJobs)
		if !flag {
			return false, msg
		}

		flag, msg = IsItemModuleResourceValid(item.ItemModules)
		if !flag {
			return false, msg
		}

		if len(item.RunAfter) == 0 {
			fatherNum++
		}

		_, ok := itemNames[item.Name]
		if ok {
			return false, fmt.Sprintf("item name %s repeated", item.Name)
		}

		var itemImpl *Item
		itemImpl = &item
		itemNames[item.Name] = itemImpl

		// father item repeated
		if fatherNum > 1 {
			return false, "father num > 0"
		}
	}

	// no father item
	if fatherNum == 0 {
		return false, "job not has father item"
	}

	// parent & extend not found
	for _, item := range job.Spec.Items {
		for _, fatherName := range item.RunAfter {
			if _, ok := itemNames[item.Name]; !ok {
				return false, fmt.Sprintf("item %s parent %s not found", item.Name, fatherName)
			}
		}

		for _, itemJob := range item.ItemJobs.Jobs {
			if itemJob.ContainerExtend != nil {
				names := JobExtendStr2Names(*itemJob.ContainerExtend)
				if len(names) < 2 || !IsExtendNamesIllegal(names, itemNames) {
					return false, fmt.Sprintf("item %s job %s container extend %s not illegal",
						item.Name, itemJob.Name, *itemJob.ContainerExtend)
				}

				extendItem := itemNames[names[0]]
				extendJob := &ItemJobTemplate{}
				for _, itemJob := range extendItem.ItemJobs.Jobs {
					if itemJob.Name == names[1] {
						impl := &itemJob
						extendJob = impl
						break
					}
				}
				if extendJob.VolcanoJobSpec != nil {
					if len(extendJob.VolcanoJobSpec.Tasks) > 1 {
						return false, fmt.Sprintf("item %s job %s container extend %s not illegal: not only one container",
							item.Name, itemJob.Name, *itemJob.ContainerExtend)
					}
					if extendJob.VolcanoJobSpec.Tasks[0].Replicas != 1 {
						return false, fmt.Sprintf("item %s job %s container extend %s not illegal: not only one container",
							item.Name, itemJob.Name, *itemJob.ContainerExtend)
					}
				} else if extendJob.KubeJobSpec != nil {
					if extendJob.KubeJobSpec.Parallelism != nil && *extendJob.KubeJobSpec.Parallelism > 0 {
						return false, fmt.Sprintf("item %s job %s container extend %s not illegal: not only one container",
							item.Name, itemJob.Name, *itemJob.ContainerExtend)
					}
				}
			}

			if itemJob.NodeNameExtend != nil {
				names := JobExtendStr2Names(*itemJob.NodeNameExtend)
				if len(names) < 2 || !IsExtendNamesIllegal(names, itemNames) {
					return false, fmt.Sprintf("item %s job %s node_name extend %s not illegal",
						item.Name, itemJob.Name, *itemJob.ContainerExtend)
				}
			}
		}
	}

	return true, ""
}

func JobExtendStr2Names(s string) []string {
	return strings.Split(s, "->")
}

func IsItemModuleResourceValid(modules ItemModuleResource) (bool, string) {

	for _, svc := range modules.Services {
		if svc.Name == "" {
			return false, fmt.Sprintf("service name can not be nil")
		}
	}

	for _, cm := range modules.ConfigMaps {
		if cm.Name == "" {
			return false, fmt.Sprintf("configmap name can not be nil")
		}
	}

	for _, sc := range modules.Secrets {
		if sc.Name == "" {
			return false, fmt.Sprintf("secret name can not be nil")
		}
	}

	for _, pvc := range modules.Pvcs {
		if pvc.Name == "" {
			return false, fmt.Sprintf("pvc name can not be nil")
		}
	}

	for _, pv := range modules.Pvs {
		if pv.Name == "" {
			return false, fmt.Sprintf("pv name can not be nil")
		}
	}

	return true, ""
}

func IsItemJobResourceValid(jobs ItemJobResource) (bool, string) {
	for _, job := range jobs.Jobs {
		if job.VolcanoJobSpec == nil && job.KubeJobSpec == nil {
			return false, fmt.Sprintf("kube_job and volcano_job can not be total nil")
		}

		if job.VolcanoJobSpec != nil && job.KubeJobSpec != nil {
			return false, fmt.Sprintf("kube_job and volcano_job can not be total set")
		}

		if job.Name == "" {
			return false, fmt.Sprintf("job name can not be nil")
		}
	}

	return true, ""
}

func IsExtendNamesIllegal(names []string, itemNames map[string]*Item) bool {

	item := &Item{}
	itemJob := &ItemJobTemplate{}

	for i, name := range names {
		switch i {
		case 0:
			ok := false
			item, ok = itemNames[name]
			if !ok {
				return false
			}
		case 1:
			for _, jobTemplate := range item.ItemJobs.Jobs {
				if jobTemplate.Name == name {
					itemJob = &jobTemplate
					break
				}
			}

			if itemJob.Name == "" {
				return false
			}

		case 2:
			if itemJob.KubeJobSpec != nil && itemJob.KubeJobSpec.Template.Name != name {
				return false
			}

			flag := false
			if itemJob.VolcanoJobSpec != nil {
				for _, task := range itemJob.VolcanoJobSpec.Tasks {
					if task.Name == name {
						flag = true
						break
					}
				}
			}

			if flag == false {
				return false
			}
		}
	}

	return true
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
