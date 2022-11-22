package main

import "fmt"

// Map 使用拉链法实现
type Map interface {
	Put(k, v int) bool
	Get(k int) (int, bool)
	Del(k int) bool
}

type OpenMap struct {
	size  int
	slots []*OpenSlotNode
}

type OpenSlotNode struct {
	Key, Val int
	Next     *OpenSlotNode
}

func (m *OpenMap) Put(k, v int) bool {
	h := m.hash(k)
	pre := m.slots[h]
	node := pre.Next
	for node != nil {
		pre = node
		if node.Key == k {
			// has existed
			node.Val = v
			return false
		}
		node = node.Next
	}
	// not existed
	pre.Next = &OpenSlotNode{k, v, nil}
	return true
}

func (m *OpenMap) Get(k int) (int, bool) {
	h := m.hash(k)
	node := m.slots[h].Next
	for node != nil {
		if node.Key == k {
			// has existed
			return node.Val, true
		}
		node = node.Next
	}
	// not existed
	return 0, false
}

func (m *OpenMap) Del(k int) bool {
	h := m.hash(k)
	var pre *OpenSlotNode
	node := m.slots[h]
	for node != nil {
		pre = node
		node = node.Next
		if node == nil {
			return false
		}
		if node.Key == k {
			// has existed
			pre.Next = node.Next
			return true
		}
	}
	// not existed
	return false
}

func (m *OpenMap) hash(k int) int {
	return k % m.size
}

type LRUSlotNode struct {
	Key, Val int
	Next     *LRUSlotNode
	DNode    *DNode
}

type DList struct {
	Head, Tail *DNode
}

func NewDList() *DList {
	head, tail := NewDNode(nil), NewDNode(nil)
	head.Next = tail
	tail.Prev = head
	return &DList{head, tail}
}

type DNode struct {
	Prev, Next  *DNode
	PreSlotNode *LRUSlotNode
}

func NewDNode(slot *LRUSlotNode) *DNode {
	return &DNode{nil, nil, slot}
}

type LRUMap struct {
	size, curr int
	slots      []*LRUSlotNode
	list       *DList
}

func (m *LRUMap) Put(k, v int) bool {
	h := m.hash(k)
	pre := m.slots[h]
	node := pre.Next
	for node != nil {
		pre = node
		if node.Key == k {
			// has existed
			node.Val = v
			m.list.MoveToHead(node.DNode)
			return false
		}
		node = node.Next
	}
	// not existed
	m.curr++
	if m.curr > m.size {
		dnode := m.list.RemoveTail()
		dnode.PreSlotNode.Next = dnode.PreSlotNode.Next.Next // del
		m.curr--
		if pre == dnode.PreSlotNode.Next {
			pre = pre.DNode.PreSlotNode
		}
	}
	pre.Next = &LRUSlotNode{k, v, nil, nil}
	dnode := &DNode{nil, nil, pre}
	m.list.AddToHead(dnode)
	pre.Next.DNode = dnode
	return true
}

func (m *LRUMap) Get(k int) (int, bool) {
	h := m.hash(k)
	node := m.slots[h].Next
	for node != nil {
		if node.Key == k {
			// has existed
			m.list.MoveToHead(node.DNode)
			return node.Val, true
		}
		node = node.Next
	}
	// not existed
	return 0, false
}

func (m *LRUMap) Del(k int) bool {
	h := m.hash(k)
	var pre *LRUSlotNode
	node := m.slots[h]
	for node != nil {
		pre = node
		node = node.Next
		if node == nil {
			return false
		}
		if node.Key == k {
			// has existed
			pre.Next = node.Next
			m.list.RemoveNode(node.DNode)
			m.curr--
			return true
		}
	}
	// not existed
	return false
}

func (m *LRUMap) hash(k int) int {
	return k % m.size
}

func (list *DList) AddToHead(node *DNode) {
	node.Next = list.Head.Next
	node.Prev = list.Head
	list.Head.Next.Prev = node
	list.Head.Next = node
}

func (list *DList) RemoveNode(node *DNode) {
	if node == nil {
		return
	}
	if node.Prev == nil {
		return
	}
	node.Prev.Next = node.Next
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}
	node.Prev = nil
	node.Next = nil
}

func (list *DList) MoveToHead(node *DNode) {
	list.RemoveNode(node)
	list.AddToHead(node)
}

func (list *DList) RemoveTail() *DNode {
	node := list.Tail.Prev
	list.RemoveNode(node)
	return node
}

type MapType int

const (
	openMap MapType = iota
	lruMap
)

func NewMap(size int, mapType MapType) Map {
	switch mapType {
	case openMap:
		slots := make([]*OpenSlotNode, size)
		for i := range slots {
			slots[i] = &OpenSlotNode{-1, -1, nil}
		}
		return &OpenMap{size, slots}
	case lruMap:
		slots := make([]*LRUSlotNode, size)
		for i := range slots {
			slots[i] = &LRUSlotNode{-1, -1, nil, nil}
		}
		return &LRUMap{size, 0, slots, NewDList()}
	}
	return nil
}

type TestCase struct {
	k, v int
}

func main() {
	testMap(openMap)
	// testMap(lruMap)
}

func testMap(mapType MapType) {
	m := NewMap(2, mapType)
	putcases := []TestCase{
		{1, 2},
		{2, 3},
		{3, 4},
		{1, 1},
		{4, 5},
	}
	for _, p := range putcases {
		/*
		   m.Put(1, 2)
		   m.Put(2, 3)
		   m.Put(3, 4) // lrumap del the (1,2)
		   m.Put(1, 1) // lrumap del the (2,3)
		   m.Put(4, 5) // lrumap del the (3,4)
		*/
		ok := m.Put(p.k, p.v)
		fmt.Printf("%v,", ok)
	}
	/*
	   openmap except: true,true,true,false,true,
	   lrumap except: true,true,true,true,true,

	*/
	fmt.Println()
	fmt.Println(m.Del(2))
	/*
	   openmap except: true
	   lrumap except: false

	*/
	for _, p := range putcases {
		v, ok := m.Get(p.k)
		fmt.Printf("%d %v,", v, ok)
	}
	/*
	   openmap except: 1 true,0 false,4 true,1 true,5 true,
	   lrumap except:  1 true,0 false,0 false,1 true,5 true,

	*/
	fmt.Println()
}
