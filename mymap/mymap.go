package mymap

import (
	"sync"
)

type Node struct {
	value      string
	prev, next *Node
}

// we implement thread safety with RWMutex where we simply allow reads to be simultaneous
// while inserting/updating exclusive. This way we can have multiple getItem and getAllItems
// commands running simultaneiosly while addItem and deleteItem running in exclusive lock.
// We could also further optimise thread safety functionality by simply locking specific
// nodes instead which take part in modifying data but not sure if we need that or this is enough.
type OrderedSet struct {
	mu          sync.RWMutex
	elementsMap map[string]*Node
	head, tail  *Node
}

func NewOrderedSet() *OrderedSet {
	return &OrderedSet{
		elementsMap: make(map[string]*Node),
		head:        &Node{},
		tail:        &Node{},
	}
}

func (s *OrderedSet) insertAfter(prev *Node, value string) *Node {
	newNode := &Node{value: value, prev: prev, next: prev.next}
	prev.next = newNode
	if newNode.next != nil {
		newNode.next.prev = newNode
	}
	return newNode
}

func (s *OrderedSet) InsertElement(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.elementsMap[value]; ok {
		return // Element already exists
	}
	if len(s.elementsMap) == 0 {
		s.head.next = s.insertAfter(s.head, value)
		s.tail.prev = s.head.next
		s.elementsMap[key] = s.head.next
	} else {
		// if exists - update
		if _, ok := s.elementsMap[key]; ok {
			s.elementsMap[key].value = value
			return
		}
		newNode := s.insertAfter(s.tail.prev, value)
		s.tail.prev = newNode
		s.elementsMap[key] = newNode
	}
}

func (s *OrderedSet) RemoveElement(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if node, ok := s.elementsMap[value]; ok {
		node.prev.next = node.next
		if node.next != nil {
			node.next.prev = node.prev
		} else {
			s.tail.prev = node.prev
		}
		delete(s.elementsMap, value)
	}
}

func (s *OrderedSet) GetElement(value string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if node, ok := s.elementsMap[value]; ok {
		return node.value
	}
	return ""
}

func (s *OrderedSet) GetAllElements() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	elements := make([]string, 0, len(s.elementsMap))
	current := s.head.next
	for current != nil && current != s.tail {
		elements = append(elements, current.value)
		current = current.next
	}
	return elements
}
