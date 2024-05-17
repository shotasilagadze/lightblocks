package mymap

import (
	"sync"
)

type Node struct {
	value      string
	prev, next *Node
}

// Implementation is using simple hash map (unordered_map in c++) implementation for O(1)
// operations. To maintain the order of the insertion we use double linked list to traverse
// upon requesting all the elements from the map.

// We implement thread safety with RWMutex where we simply allow reads to be simultaneous
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
	orderedSet := &OrderedSet{
		elementsMap: make(map[string]*Node),
		head:        &Node{},
		tail:        &Node{},
	}

	// connect head and tail
	orderedSet.head.next = orderedSet.tail
	orderedSet.tail.prev = orderedSet.head

	return orderedSet
}

func (s *OrderedSet) insertAfter(prev *Node, value string) *Node {
	prev.next = &Node{value: value, prev: prev, next: prev.next}
	prev.next.next.prev = prev.next
	return prev.next
}

func (s *OrderedSet) InsertElement(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.elementsMap) == 0 {
		s.elementsMap[key] = s.insertAfter(s.head, value)
		return
	}

	// if exists - update
	if node, ok := s.elementsMap[key]; ok {
		node.value = value
		return
	}

	s.elementsMap[key] = s.insertAfter(s.tail.prev, value)
}

func (s *OrderedSet) RemoveElement(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if node, ok := s.elementsMap[key]; ok {
		node.prev.next = node.next
		node.next.prev = node.prev
		delete(s.elementsMap, key)
	}
}

func (s *OrderedSet) GetElement(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if node, ok := s.elementsMap[key]; ok {
		return node.value
	}
	return ""
}

func (s *OrderedSet) GetAllElements() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	elements := make([]string, 0, len(s.elementsMap))
	current := s.head.next
	for current != s.tail {
		elements = append(elements, current.value)
		current = current.next
	}
	return elements
}
