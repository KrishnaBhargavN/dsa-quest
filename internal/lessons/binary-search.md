---
title: Binary search
group: Foundations
order: 6
topics: Binary search
minutes: 15
summary: Binary search finds the boundary of any yes/no question that flips exactly once, on an array or on the answer itself.
---

## How to recognise it

- Sorted (or rotated sorted) input and an O(log n) requirement.
- "Minimum x such that…", "maximum x such that…", "smallest capacity / speed / largest minimum".
- You can easily **check** a candidate answer, but constructing the best one directly is hard.
- The answer range is huge (up to 10^9), far too big to try every value.

## The one idea behind every version

Binary search does not need a sorted array. It needs a **predicate** that is false, false, …, false, true, true, …, true over the search range (monotonic). It finds the first true. Sorted arrays are just one source of such a predicate: "a[i] ≥ target" flips once.

## The template (first true)

```java
// Finds the smallest x in [lo, hi] with ok(x) true; returns hi + 1 if none.
long lo = LOW, hi = HIGH + 1;          // search in the half-open range [lo, hi)
while (lo < hi) {
    long mid = lo + (hi - lo) / 2;
    if (ok(mid)) hi = mid;             // mid might be the answer; keep it
    else lo = mid + 1;                 // mid is definitely too small
}
return lo;
```

**Why it terminates and is correct:**
- Invariant: everything below lo is false, everything at or above hi is true (treat hi as true).
- Each step shrinks [lo, hi) and keeps the invariant: if ok(mid), the first true is at mid or before, so hi = mid; otherwise it's after mid.
- When lo == hi the invariant pins the boundary exactly.
- `lo + (hi - lo) / 2` avoids overflow, and since mid < hi the range always shrinks.

Use this single template for everything. For "last true", search for the first index where the predicate is **false** and subtract one.

## On sorted arrays

- Lower bound (first a[i] ≥ x): ok(i) = a[i] ≥ x.
- Upper bound (first a[i] > x): ok(i) = a[i] > x.
- Count of x = upper − lower.
- Java's `Arrays.binarySearch` returns **some** matching index, not the first, so don't rely on it for boundaries.

## Rotated sorted arrays

At any mid, one half is sorted: if a[lo] ≤ a[mid], the left half is sorted, otherwise the right half is. Check whether the target lies inside the sorted half's range; if yes go there, else go to the other half.

For the minimum, compare a[mid] with a[hi]: if a[mid] > a[hi], the drop (and the minimum) is strictly right of mid; otherwise the minimum is at mid or left of it. Comparing with a[hi] (not a[lo]) avoids the unrotated special case.

## Binary search on the answer

This is the big one for interviews. Template thought: "If I fix the answer to X, can I check feasibility greedily in O(n)?"

- **Koko Eating Bananas:** speed s works if Σ ceil(p / s) ≤ h. Faster always works if slower works, so ok(s) is monotonic. Search s in [1, max pile].
- **Ship Packages in D Days:** capacity c works if a greedy fill uses ≤ D days. Range [max weight, total].
- **Split Array Largest Sum:** same as shipping, with k parts.
- **Bouquets:** "can I make m bouquets by day D" is monotonic in D.
- **Minimise the maximum / maximise the minimum** problems are almost always this.

**Why the greedy check is right (shipping):** for a fixed capacity, loading each day as full as possible never uses more days than any other loading. Exchange argument: in any valid loading, moving the first package of day 2 into day 1 (if it fits) can't increase the number of days. Repeating that turns any loading into the greedy one.

## Median of two sorted arrays

Partition A at i and B at j = (m + n + 1) / 2 − i so the left side has half the elements. The partition is correct when A[i−1] ≤ B[j] and B[j−1] ≤ A[i]. If A[i−1] > B[j], i is too big; otherwise too small. Search i over the **shorter** array: O(log min(m, n)). Use ±infinity for out-of-range indices.

## Real-valued binary search

For doubles, loop a fixed number of times (100 iterations is plenty) instead of `while (lo < hi)`, which may never end with floating point.

## Common bugs

- Mixing templates (`lo <= hi` with `hi = mid`) causes infinite loops. Use the one above every time.
- Overflow in `lo + hi`. Use `lo + (hi - lo) / 2` and `long` bounds.
- A non-monotonic predicate. Test it mentally: "if x works, does x + 1 work?"
- Wrong range: the lower bound must be a value that could be the answer (for shipping, max weight, not 1).

## Where it shows up

- **Search in Rotated Sorted Array, Find Minimum in Rotated Sorted Array.**
- **Time Based Key-Value Store:** last timestamp ≤ t (or `TreeMap.floorEntry`).
- **Koko, Ship Packages, Bouquets, Split Array Largest Sum:** search on the answer.
- **Median of Two Sorted Arrays:** search on a partition.
- **Swim in Rising Water:** search on time plus a BFS check (one of three valid approaches).
- **Longest Increasing Subsequence in O(n log n):** lower bound into the tails array.
