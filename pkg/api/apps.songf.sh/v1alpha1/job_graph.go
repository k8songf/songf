package v1alpha1

import "fmt"

type ItemNode struct {
	Item  *Item
	Child []*ItemNode
}

func NewGraphFromJob(job *Job) (*ItemNode, map[string]*ItemNode, error) {

	nodeMap := map[string]*ItemNode{}
	var fatherNode *ItemNode

	for _, item := range job.Spec.Items {
		if _, ok := nodeMap[item.Name]; ok {
			return nil, nil, fmt.Errorf("new job item tree build err: item Name %s repeated", item.Name)
		}

		var itemImpl *Item
		itemImpl = &item
		nodeMap[item.Name] = &ItemNode{
			Item: itemImpl,
		}

		if len(item.RunAfter) == 0 {
			if fatherNode != nil {
				return nil, nil, fmt.Errorf("new job item tree build err: muilty start item")
			} else {
				fatherNode = nodeMap[item.Name]
			}
		}
	}

	for _, node := range nodeMap {
		var nodeImpl *ItemNode
		nodeImpl = node

		for _, parentName := range node.Item.RunAfter {
			if _, ok := nodeMap[parentName]; !ok {
				return nil, nil, fmt.Errorf("not find parent item Name %s", parentName)
			} else {
				nodeMap[parentName].Child = append(nodeMap[parentName].Child, nodeImpl)
			}
		}
	}

	return fatherNode, nodeMap, nil
}

func (n *ItemNode) Children() []*ItemNode {
	return n.Child
}

func (n *ItemNode) IsFirst() bool {
	if len(n.Item.RunAfter) == 0 {
		return true
	}

	return false
}

func (n *ItemNode) isLast() bool {
	if len(n.Child) == 0 {
		return true
	}

	return false
}
