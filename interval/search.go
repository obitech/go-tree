package interval

import "fmt"

const noIntervalErrMsg = "no interval found for %q"

type inorderResult struct {
	nodes   map[Interval]*node
	results []Result
}

// FindFirstOverlapping returns the payload of the first interval that overlaps
// with the passed key. Returns an ErrNotFound if no overlapping interval is
// found.
func (t *Tree) FindFirstOverlapping(key Interval) (Result, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.root == t.sentinel {
		return Result{}, ErrNotFound(fmt.Sprintf(noIntervalErrMsg, key))
	}

	n := t.search(t.root, key)

	if n == t.sentinel {
		return Result{}, ErrNotFound(fmt.Sprintf(noIntervalErrMsg, key))
	}

	return Result{
		Interval: n.key,
		Payload:  n.payload,
	}, nil
}

// FindAllOverlapping returns a slice of Result with all intervals overlapping
// the given interval key. Returns an ErrNotFound if no overlapping interval is
// found.
func (t *Tree) FindAllOverlapping(key Interval) ([]Result, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.root == t.sentinel {
		return nil, ErrNotFound(fmt.Sprintf(noIntervalErrMsg, key))
	}

	res := &inorderResult{
		nodes:   make(map[Interval]*node),
		results: make([]Result, 0),
	}

	t.searchInorder(t.root, key, res)

	if len(res.results) == 0 {
		return nil, ErrNotFound(fmt.Sprintf(noIntervalErrMsg, key))
	}

	return res.results, nil
}

// FindExact returns the exactly matching Result for the given key interval.
// Returns an ErrNotFound if not found.
func (t *Tree) FindExact(key Interval) (Result, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if n := t.findExact(key); n != nil {
		return Result{
			Interval: n.key,
			Payload:  n.payload,
		}, nil
	}

	return Result{}, ErrNotFound(fmt.Sprintf("interval %q does not exist", key))
}

// InOrder returns an ordered list of all entries.
func (t *Tree) InOrder() []Result {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.root == t.sentinel {
		return nil
	}

	res := make([]Result, 0)

	t.resultsInorder(t.root, &res)

	return res
}

func (t *Tree) resultsInorder(z *node, res *[]Result) {
	if z == t.sentinel {
		return
	}

	if z.left != t.sentinel {
		t.resultsInorder(z.left, res)
	}

	*res = append(*res, Result{
		Interval: z.key,
		Payload:  z.payload,
	})

	if z.right != t.sentinel {
		t.resultsInorder(z.right, res)
	}
}

func (t *Tree) findExact(key Interval) *node {
	if t.root == t.sentinel {
		return nil
	}

	res := &inorderResult{
		nodes:   make(map[Interval]*node),
		results: make([]Result, 0),
	}

	t.searchInorder(t.root, key, res)

	if n, ok := res.nodes[key]; ok {
		return n
	}

	return nil
}

func (t *Tree) searchInorder(z *node, key Interval, result *inorderResult) {
	if result == nil {
		panic("result can't be nil")
	}

	if z.left != t.sentinel && greaterOrEqual(z.left.max, key.low) {
		t.searchInorder(z.left, key, result)
	}

	if z.key.overlaps(key) {
		result.nodes[z.key] = z
		result.results = append(result.results, Result{
			Interval: z.key,
			Payload:  z.payload,
		})
	}

	if z.right != t.sentinel && greaterOrEqual(z.right.max, key.low) {
		t.searchInorder(z.right, key, result)
	}
}

func (t *Tree) search(x *node, key Interval) *node {
	for x != t.sentinel && !key.overlaps(x.key) {
		if x.left != t.sentinel && greaterOrEqual(x.left.max, key.low) {
			x = x.left
		} else {
			x = x.right
		}
	}

	return x
}
