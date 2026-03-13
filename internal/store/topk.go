package store

import (
	"container/heap"
	"sort"
)

// scored pairs a Record with its similarity score.
type scored struct {
	rec   Record
	score float64
}

// minHeap is a min-heap of scored records keyed on score.
// It keeps the k highest-scored records seen so far, evicting the lowest
// when full — giving O(n log k) top-k selection vs O(n log n) for a full sort.
type minHeap []scored

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].score < h[j].score } // min at root
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)         { *h = append(*h, x.(scored)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopK returns up to k records nearest to vec by cosine similarity,
// ordered highest score first.
func TopK(records []Record, vec []float32, k int) []Record {
	if k <= 0 || len(records) == 0 {
		return nil
	}
	h := &minHeap{}
	heap.Init(h)

	for _, r := range records {
		sc := Cosine(vec, r.Embedding)
		if h.Len() < k {
			heap.Push(h, scored{rec: r, score: sc})
		} else if sc > (*h)[0].score {
			heap.Pop(h)
			heap.Push(h, scored{rec: r, score: sc})
		}
	}

	// Sort heap contents descending before returning.
	sort.Slice(*h, func(i, j int) bool { return (*h)[i].score > (*h)[j].score })

	out := make([]Record, h.Len())
	for i, s := range *h {
		out[i] = s.rec
	}
	return out
}
