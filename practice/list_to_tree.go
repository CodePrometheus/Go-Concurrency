package main

import (
	"encoding/json"
	"log"
)

type Tree struct {
	List     map[int]*Node
	Children map[int]Node
	Parents  map[int]Node
}

type Node struct {
	Id    int     `json:"id"`
	Pid   int     `json:"pid"`
	Child []*Node `json:"child"`
}

func (this *Tree) build(nodes []Node) {
	this.List = make(map[int]*Node, 0)
	for index, _ := range nodes {
		id := nodes[index].Id
		nodes[index].Child = make([]*Node, 0)
		this.List[id] = &nodes[index]
	}
	for k, _ := range this.List {
		pid := this.List[k].Pid
		if _, ok := this.List[pid]; ok {
			this.List[pid].Child = append(this.List[pid].Child, this.List[k])
		}
	}
	for k, _ := range this.List {
		if this.List[k].Id > 1 {
			delete(this.List, k)
		}
	}
}

// GetAllNode ... 获取所有子节点
func GetAllNode(node *Node) (nodes []string) {
	if len(node.Child) == 0 {
		nodes = append(nodes)
		return nodes
	}
	for _, t := range node.Child {
		for _, n := range GetAllNode(t) {
			nodes = append(nodes, n)
		}
	}
	return nodes
}
func main() {
	menus := []byte(
		`[{"id":1,"pid":0},{"id":3,"pid":1},{"id":2,"pid":1},
	{"id":4,"pid":2},{"id":5,"pid":3},{"id":6,"pid":2}]`)
	TestNode(menus)
}

func TestNode(menus []byte) {
	var nodes []Node
	err := json.Unmarshal(menus, &nodes)
	if err != nil {
		log.Fatal("JSON decode error:", err)
		return
	}
	// 构建树
	var exampleTree Tree
	exampleTree.build(nodes)
	bs, _ := json.Marshal(exampleTree.List)
	log.Println("tree:", string(bs))
	// 获取节点1的所有子节点
	n := GetAllNode(exampleTree.List[1])
	log.Println("n:", n)
}
