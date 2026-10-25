---
title: Range queries: Fenwick and segment trees
group: Advanced
order: 30
topics: Range queries
minutes: 20
summary: Answer "sum / min / max over [l, r]" while the array keeps changing, in O(log n) per operation. Fenwick for sums, segment trees for everything else, lazy propagation for range updates.
---

## How to recognise it

- Many range queries **and** updates interleaved. (Queries only, no updates: prefix sums or a sparse table are enough.)
- "Count smaller elements before / after", inversions, ranks over a changing set.
- "First index where the value is at least x" with updates.
- n and q up to 2 × 10^5.

## Choosing

| Need | Use |
| --- | --- |
| Range sum, point update | Fenwick tree |
| Range add, point query | Fenwick on the difference array |
| Range add, range sum | Two Fenwicks, or a lazy segment tree |
| Range min/max/gcd, point update | Segment tree |
| Range min, no updates | Sparse table (O(1) query) |
| Range assign/add + range min/sum | Lazy segment tree |
| "First index with value ≥ x" | Segment tree descent |

## Fenwick tree (binary indexed tree)

```java
long[] bit = new long[n + 1];                     // 1-indexed
void add(int i, long delta) { for (; i <= n; i += i & -i) bit[i] += delta; }
long prefix(int i) { long s = 0; for (; i > 0; i -= i & -i) s += bit[i]; return s; }
long range(int l, int r) { return prefix(r) - prefix(l - 1); }
```

**Why it works:** `i & -i` is the lowest set bit of i. bit[i] stores the sum of the block of length lowbit(i) ending at i. Any prefix [1, i] splits into at most log n such blocks (one per set bit of i), which `prefix` walks by stripping the lowest bit each time. An update at i must touch every block that contains i; those are exactly i, i + lowbit(i), and so on, which `add` walks. Both loops run O(log n) times.

Point **set** (not add): keep the array and add (newValue − oldValue).

**Range add, point query:** apply the difference-array trick inside the Fenwick: add(l, v), add(r + 1, −v); the value at k is original[k] + prefix(k).

**Counting with a Fenwick over values (Count of Smaller Numbers After Self):** compress values to ranks 1..m. Scan from the right: the answer for x is prefix(rank(x) − 1) (how many smaller values were seen), then add(rank(x), 1).

## Segment tree (iterative, point update)

For min, max, gcd or any associative operation:

```java
int size; long[] t;                               // leaves at t[size .. 2*size)
void build(long[] a) {
    size = a.length; t = new long[2 * size];
    for (int i = 0; i < size; i++) t[size + i] = a[i];
    for (int i = size - 1; i > 0; i--) t[i] = Math.min(t[2 * i], t[2 * i + 1]);
}
void set(int i, long v) {
    for (t[i += size] = v; i > 1; i >>= 1) t[i >> 1] = Math.min(t[i], t[i ^ 1]);
}
long query(int l, int r) {                        // inclusive l..r
    long res = Long.MAX_VALUE;
    for (l += size, r += size + 1; l < r; l >>= 1, r >>= 1) {
        if ((l & 1) == 1) res = Math.min(res, t[l++]);
        if ((r & 1) == 1) res = Math.min(res, t[--r]);
    }
    return res;
}
```

**Why a query is O(log n):** each node covers a contiguous block; any range is the union of at most two nodes per level, which the loop collects while climbing.

## Segment tree descent (Hotel Queries)

To find the **first** index with value ≥ x, use a recursive max segment tree built over a power-of-two size. From the root, go left if the left child's max ≥ x, otherwise go right. If the root's max < x, there is none. You visit one node per level: O(log n). Then update that leaf and recompute its ancestors.

## Lazy propagation (range updates)

Range updates like "add v to [l, r]" would touch O(n) leaves. Instead, when a node's range is fully inside [l, r], update the node's aggregate directly and store a **pending tag** ("my children still owe +v"). Before you ever go below a node (in a later update or query), **push** its tag down to its two children.

```java
long[] mn = new long[4 * n], lazy = new long[4 * n];
void apply(int node, long v) { mn[node] += v; lazy[node] += v; }
void push(int node) {
    if (lazy[node] != 0) { apply(2 * node, lazy[node]); apply(2 * node + 1, lazy[node]); lazy[node] = 0; }
}
void update(int node, int lo, int hi, int l, int r, long v) {
    if (r < lo || hi < l) return;
    if (l <= lo && hi <= r) { apply(node, v); return; }
    push(node);
    int mid = (lo + hi) / 2;
    update(2 * node, lo, mid, l, r, v);
    update(2 * node + 1, mid + 1, hi, l, r, v);
    mn[node] = Math.min(mn[2 * node], mn[2 * node + 1]);
}
long query(int node, int lo, int hi, int l, int r) {
    if (r < lo || hi < l) return Long.MAX_VALUE;
    if (l <= lo && hi <= r) return mn[node];
    push(node);
    int mid = (lo + hi) / 2;
    return Math.min(query(2 * node, lo, mid, l, r), query(2 * node + 1, mid + 1, hi, l, r));
}
```

**Why it's correct:** the invariant is "a node's value is right for its whole range, assuming its own pending tag has been applied; its children may be stale by exactly that tag". Applying a tag to a node updates its value immediately; pushing before descending makes the children correct before you rely on them. Each operation touches O(log n) nodes, like a query.

For range **sum** with range add, the node's value increases by v × (number of elements in the node), so `apply` needs the node's length. For range **assign**, the tag replaces rather than adds, and you need a "no tag" marker.

## Sparse table (static range min)

st[k][i] = min of a[i .. i + 2^k − 1], built in O(n log n). A query [l, r] overlaps two blocks of length 2^k (k = floor(log2(r − l + 1))): min(st[k][l], st[k][r − 2^k + 1]). Overlap is fine for min/max (idempotent), not for sums.

## Common bugs

- Mixing 0- and 1-indexing in a Fenwick (index 0 loops forever in `add`).
- `int` overflow in sums: use `long`.
- Forgetting to push before descending in a lazy tree.
- Recursive segment tree arrays sized 2n instead of 4n.
- Slow Java I/O on CSES (see the Knapsack lesson).

## Where it shows up

- **Range Sum Query - Mutable, Count of Smaller Numbers After Self, Dynamic Range Sum Queries, Dynamic Range Minimum Queries, Range Update Queries, Hotel Queries.**
- Lazy propagation isn't on your list but shows up in harder rounds ("range add, range min"); try CSES Range Updates and Sums once you're comfortable.
