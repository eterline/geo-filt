package somedata

type DisjointNumeric interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64
}

type DisjointSet[T comparable] interface {
	MakeSet(x T)
	Find(x T) T
	Union(x, y T)
}

// =======================

type arrDisjointSet[T DisjointNumeric] struct {
	parent []T
	rank   []T
}

func (ds *arrDisjointSet[T]) MakeSet(x T) {
	ds.parent[int(x)] = x
	ds.rank[int(x)] = 0
}

func (ds *arrDisjointSet[T]) Find(x T) T {
	if ds.parent[int(x)] != x {
		ds.parent[int(x)] = ds.Find(ds.parent[int(x)])
	}
	return ds.parent[int(x)]
}

func (ds *arrDisjointSet[T]) Union(x, y T) {
	xRoot := ds.Find(x)
	yRoot := ds.Find(y)
	if xRoot == yRoot {
		return
	}
	if ds.rank[int(xRoot)] < ds.rank[int(yRoot)] {
		ds.parent[int(xRoot)] = yRoot
	} else if ds.rank[int(xRoot)] > ds.rank[int(yRoot)] {
		ds.parent[int(yRoot)] = xRoot
	} else {
		ds.parent[int(yRoot)] = xRoot
		ds.rank[int(xRoot)]++
	}
}

type mapDisjointSet[T comparable] struct {
	parent map[T]T
	rank   map[T]int
}

// =======================

func (ds *mapDisjointSet[T]) MakeSet(x T) {
	if _, ok := ds.parent[x]; !ok {
		ds.parent[x] = x
		ds.rank[x] = 0
	}
}

func (ds *mapDisjointSet[T]) Find(x T) T {
	if ds.parent[x] != x {
		ds.parent[x] = ds.Find(ds.parent[x]) // path compression
	}
	return ds.parent[x]
}

func (ds *mapDisjointSet[T]) Union(x, y T) {
	xRoot := ds.Find(x)
	yRoot := ds.Find(y)
	if xRoot == yRoot {
		return
	}
	if ds.rank[xRoot] < ds.rank[yRoot] {
		ds.parent[xRoot] = yRoot
	} else if ds.rank[xRoot] > ds.rank[yRoot] {
		ds.parent[yRoot] = xRoot
	} else {
		ds.parent[yRoot] = xRoot
		ds.rank[xRoot]++
	}
}
