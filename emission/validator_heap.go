// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import "container/heap"

// ValidatorHeapItem wraps Validator with an index for heap.Interface.
type ValidatorHeapItem struct {
	validator *Validator
	index     int // The index of the item in the heap.
}

// ValidatorHeap implements heap.Interface and holds ValidatorHeapItems.
type ValidatorHeap []*ValidatorHeapItem

func (vh ValidatorHeap) Len() int { return len(vh) }

func (vh ValidatorHeap) Less(i, j int) bool {
	// This example uses total stake as the comparison metric. Adjust as needed.
	return vh[i].validator.StakedAmount+vh[i].validator.DelegatedAmount < vh[j].validator.StakedAmount+vh[j].validator.DelegatedAmount
}

func (vh ValidatorHeap) Swap(i, j int) {
	vh[i], vh[j] = vh[j], vh[i]
	vh[i].index = i
	vh[j].index = j
}

func (vh *ValidatorHeap) Push(x interface{}) {
	n := len(*vh)
	item := x.(*ValidatorHeapItem)
	item.index = n
	*vh = append(*vh, item)
}

func (vh *ValidatorHeap) Pop() interface{} {
	old := *vh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // Avoid memory leak
	item.index = -1 // For safety
	*vh = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the heap.
func (vh *ValidatorHeap) update(item *ValidatorHeapItem, validator *Validator) {
	item.validator = validator
	heap.Fix(vh, item.index)
}
