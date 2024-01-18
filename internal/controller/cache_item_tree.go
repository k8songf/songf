package controller

import (
	"fmt"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	alpha1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

type jobItemTree struct {
	name      string
	uuid      types.UID
	nameSpace string

	startItemNode *itemNode

	workNodes map[string]*itemNode

	itemStatus map[string]v1alpha1.ItemStatus
}

func newJobItemTree(job *v1alpha1.Job) (*jobItemTree, error) {

	nodeMap := map[string]*itemNode{}
	var fatherNode *itemNode
	itemStatus := map[string]v1alpha1.ItemStatus{}

	for _, item := range job.Spec.Items {
		if _, ok := nodeMap[item.Name]; ok {
			return nil, fmt.Errorf("new job item tree build err: item name %s repeated", item.Name)
		}

		var itemImpl *v1alpha1.Item
		itemImpl = &item
		nodeMap[item.Name] = &itemNode{
			item: itemImpl,
		}

		if len(item.RunAfter) == 0 {
			if fatherNode != nil {
				return nil, fmt.Errorf("new job item tree build err: muilty start item")
			} else {
				fatherNode = nodeMap[item.Name]
			}
		}

		flag := len(job.Status.ItemStatus) == 0
		if _, ok := job.Status.ItemStatus[item.Name]; !ok {
			flag = true
		}

		if flag {
			itemStatus[item.Name] = v1alpha1.ItemStatus{
				Name:  item.Name,
				Phase: v1alpha1.ItemPending,
			}
		} else {
			itemStatus[item.Name] = job.Status.ItemStatus[item.Name]
		}
	}

	for _, node := range nodeMap {
		var nodeImpl *itemNode
		nodeImpl = node

		for _, parentName := range node.item.RunAfter {
			if _, ok := nodeMap[parentName]; !ok {
				return nil, fmt.Errorf("not find parent item name %s", parentName)
			} else {
				nodeMap[parentName].child = append(nodeMap[parentName].child, nodeImpl)
			}
		}
	}

	tree := &jobItemTree{
		name:      job.Name,
		uuid:      job.UID,
		nameSpace: job.Namespace,

		startItemNode: fatherNode,
		workNodes:     nodeMap,
		itemStatus:    itemStatus,
	}

	if tree.hasCycle() {
		return nil, fmt.Errorf("new job item tree build err: items tree has cycle")
	}

	return tree, nil
}

func (t *jobItemTree) hasCycle() bool {
	return t.hasCycleDfs(t.startItemNode, map[string]bool{})
}

func (t *jobItemTree) hasCycleDfs(node *itemNode, visited map[string]bool) bool {

	if visited[node.item.Name] {
		return true
	}

	visited[node.item.Name] = true

	for _, child := range node.child {
		if t.hasCycleDfs(child, visited) {
			return true
		}
	}

	visited[node.item.Name] = false

	return false
}

func (t *jobItemTree) syncFromService(service *corev1.Service) error {

}

func (t *jobItemTree) syncFromKubeJob(job *v1.Job) error {

}

func (t *jobItemTree) syncFromVcJob(job *alpha1.Job) error {

}

func (t *jobItemTree) syncFromConfigmap(configmap *corev1.ConfigMap) error {

}

func (t *jobItemTree) syncFromSecret(secret *corev1.Secret) error {

}

type itemNode struct {
	item  *v1alpha1.Item
	child []*itemNode
}

func (n *itemNode) children() []*itemNode {
	return n.child
}

func (n *itemNode) isFirst() bool {
	if len(n.item.RunAfter) == 0 {
		return true
	}

	return false
}

func (n *itemNode) isLast() bool {
	if len(n.child) == 0 {
		return true
	}

	return false
}
