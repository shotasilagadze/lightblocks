package mymap

import (
	"sort"
	"strconv"
	"sync"
	"testing"
)

func TestOrderedSet_Insertion(t *testing.T) {
	orderedSet := NewOrderedSet()

	orderedSet.InsertElement("one", "one_value")
	orderedSet.InsertElement("two", "two_value")
	orderedSet.InsertElement("three", "three_value")

	elements := orderedSet.GetAllElements()
	expected := []string{"one_value", "two_value", "three_value"}
	if !equalSlices(elements, expected) {
		t.Errorf("Insertion failed. Expected %v, got %v", expected, elements)
	}
}

func TestOrderedSet_Retrieval(t *testing.T) {
	orderedSet := NewOrderedSet()

	orderedSet.InsertElement("one", "one_value")
	orderedSet.InsertElement("two", "two_value")

	if value := orderedSet.GetElement("two"); value != "two_value" {
		t.Errorf("Retrieval failed. Expected %v, got %v", "two", value)
	}
	if value := orderedSet.GetElement("four"); value != "" {
		t.Errorf("Retrieval failed. Expected %v, got %v", "", value)
	}
}

func TestOrderedSet_Removal(t *testing.T) {
	orderedSet := NewOrderedSet()

	orderedSet.InsertElement("one", "one_value")
	orderedSet.InsertElement("two", "two_value")
	orderedSet.InsertElement("three", "three_value")

	orderedSet.RemoveElement("two")

	elements := orderedSet.GetAllElements()
	expected := []string{"one_value", "three_value"}
	if !equalSlices(elements, expected) {
		t.Errorf("Removal failed. Expected %v, got %v", expected, elements)
	}
}

func TestOrderedSet_ConcurrentAccess(t *testing.T) {
	orderedSet := NewOrderedSet()

	var wg sync.WaitGroup
	numRoutines := 10
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			orderedSet.InsertElement("five", "five_value")
			orderedSet.RemoveElement("three")
			wg.Done()
		}()
	}
	wg.Wait()

	elements := orderedSet.GetAllElements()
	expected := []string{"five_value"}
	if !equalSlices(elements, expected) {
		t.Errorf("Concurrent access failed. Expected %v, got %v", expected, elements)
	}
}

func TestOrderedSet_ConcurrentInsertion(t *testing.T) {
	orderedSet := NewOrderedSet()

	var wg sync.WaitGroup
	numRoutines := 10
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(val string) {
			orderedSet.InsertElement(val, val+"_value")
			wg.Done()
		}(strconv.Itoa(i))
	}
	wg.Wait()

	elements := orderedSet.GetAllElements()
	expected := make([]string, numRoutines)
	for i := 0; i < numRoutines; i++ {
		expected[i] = strconv.Itoa(i) + "_value"
	}
	sort.Strings(elements) // Sort for comparison since insertion order is not guaranteed
	if !equalSlices(elements, expected) {
		t.Errorf("Concurrent insertion failed. Expected %v, got %v", expected, elements)
	}
}

func TestOrderedSet_ConcurrentRemoval(t *testing.T) {
	orderedSet := NewOrderedSet()

	for i := 0; i < 10; i++ {
		orderedSet.InsertElement(strconv.Itoa(i), strconv.Itoa(i)+"_value")
	}

	var wg sync.WaitGroup
	numRoutines := 5
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(val string) {
			orderedSet.RemoveElement(val)
			wg.Done()
		}(strconv.Itoa(i))
	}
	wg.Wait()

	elements := orderedSet.GetAllElements()
	expected := []string{"5_value", "6_value", "7_value", "8_value", "9_value"}
	if !equalSlices(elements, expected) {
		t.Errorf("Concurrent removal failed. Expected %v, got %v", expected, elements)
	}
}

func TestOrderedSet_MixedOperations(t *testing.T) {
	orderedSet := NewOrderedSet()

	for i := 0; i < 10; i++ {
		orderedSet.InsertElement(strconv.Itoa(i), strconv.Itoa(i)+"_value")
	}

	var wg sync.WaitGroup
	numRoutines := 10
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(val string, remove bool) {
			if remove {
				orderedSet.RemoveElement(val)
			} else {
				orderedSet.InsertElement(val, val+"_value")
			}
			wg.Done()
		}(strconv.Itoa(i), i%2 == 0)
	}
	wg.Wait()

	elements := orderedSet.GetAllElements()
	expected := []string{"1_value", "3_value", "5_value", "7_value", "9_value"}
	if !equalSlices(elements, expected) {
		t.Errorf("Mixed operations failed. Expected %v, got %v", expected, elements)
	}
}

func equalSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
