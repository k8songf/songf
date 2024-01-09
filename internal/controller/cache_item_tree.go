package controller

import (
	"fmt"
	"songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
)

type jobItemTree struct {
	startItemNode *itemNode

	workNodes map[string]*itemNode

	itemStatus map[string]v1alpha1.ItemStatus
}

func newJobItemTree(job *v1alpha1.Job) (*jobItemTree, error) {

	nodeMap := map[string]*itemNode{}
	var fatherNode *itemNode

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
		startItemNode: fatherNode,
		workNodes:     nodeMap,
		itemStatus:    map[string]v1alpha1.ItemStatus{},
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
