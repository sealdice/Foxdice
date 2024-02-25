package draw

import (
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
)

type Item[T any] struct {
	weight int
	data   T
}

type Pool[T any] struct {
	pool []Item[T]
	// 总值
	sum int
	// 区间
	totals []int
}

func BuildPool[T any](items ...Item[T]) (int, []int, []Item[T]) {
	totals := make([]int, len(items))
	num := 0
	for i, c := range items {
		weight := c.weight
		num += weight
		totals[i] = num
	}
	return num, totals, items
}

func (p *Pool[T]) Pick(shuffle bool) T {
	r := rand.Intn(p.sum) + 1
	i := find(p.totals, r)
	res := p.pool[i]
	if shuffle {
		if p.pool[i].weight-1 <= 0 {
			p.pool = slices.Delete(p.pool, i, i+1)
		} else {
			p.pool[i].weight -= 1
		}
		p.sum, p.totals, p.pool = BuildPool(p.pool...)
	}
	return res.data
}

func find(a []int, x int) int {
	i, j := 0, len(a)
	for i < j {
		h := int(uint(i+j) >> 1)
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
