package models

import (
	"math/rand"
	"sort"
)

type WeightedRandomThingGenerator[T any] struct {
    items   []Item[T]
    ready   bool
    totalWeight int
}

type Item[T any] struct {
    Value   T
    Weight  int
}

func NewWeightedRandomThingGenerator[T any]() *WeightedRandomThingGenerator[T] {
    return &WeightedRandomThingGenerator[T]{}
}

func (w *WeightedRandomThingGenerator[T]) Add(value T, weight int) {
    if w.ready {
        panic("WeightedRandomThingGenerator: cannot add after use")
    }
    w.items = append(w.items, Item[T]{Value: value, Weight: weight})
    w.totalWeight += weight
}

func (w *WeightedRandomThingGenerator[T]) RandomThing(rng *rand.Rand) T {
    if !w.ready {
        w.prepare()
    }

    target := rng.Intn(w.totalWeight)

    // Modified binary search, as Go's sort library works on slices
    i := sort.Search(len(w.items), func(i int) bool {
        cumWeight := 0
        for j := 0; j <= i; j++ {
            cumWeight += w.items[j].Weight
        }
        return cumWeight > target
    })

    // Edge case handling (rare if weights add up correctly)
    if i >= len(w.items) {
        return w.items[len(w.items)-1].Value 
    }

    return w.items[i].Value
}

func (w *WeightedRandomThingGenerator[T]) prepare() {
    // Ensure weights are sorted in ascending order for the search to work
    sort.Slice(w.items, func(i, j int) bool {
        cumWeightI := 0
        for k := 0; k <= i; k++ {
            cumWeightI += w.items[k].Weight
        }

        cumWeightJ := 0
        for k := 0; k <= j; k++ {
            cumWeightJ += w.items[k].Weight
        }

        return cumWeightI < cumWeightJ
    })
    w.ready = true
}