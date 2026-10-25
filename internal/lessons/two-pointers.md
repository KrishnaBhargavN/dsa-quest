---
title: Two pointers
group: Foundations
order: 4
topics: Two pointers
minutes: 11
summary: Two indices that only move forward (or toward each other) replace a nested loop, because each move safely rules out a whole set of candidates.
---

## How to recognise it

- The array is sorted, or sorting doesn't hurt.
- You need a pair or triple with some sum or relationship.
- In-place rearranging: remove duplicates, move zeros, partition.
- Linked lists: fast and slow pointers (see the Linked list lesson).

## Pattern 1: opposite ends

Sorted array, find a pair summing to target:

```java
int lo = 0, hi = n - 1;
while (lo < hi) {
    int s = a[lo] + a[hi];
    if (s == target) return new int[]{lo, hi};
    if (s < target) lo++; else hi--;
}
```

**Why it never skips the answer:** suppose a[lo] + a[hi] < target. Then a[lo] plus **any** element at or below hi is also too small, since the array is sorted. So a[lo] can't be in any remaining pair, and discarding it (lo++) loses nothing. The symmetric argument justifies hi−−. Each step removes one candidate for good, so there are at most n steps.

This "discard a candidate that provably can't be in the answer" argument is the heart of every two-pointer proof. When you design one, say which candidate you discard and why.

## 3Sum and k-Sum

Fix the first element i, then run opposite-end two pointers on i+1..n−1 for target −a[i]. O(n^2).

Avoiding duplicate triplets: since the array is sorted, equal values sit together.
- Skip i if a[i] == a[i−1].
- After recording a match, move lo past all equal values and hi past all equal values.

## Container With Most Water

Area = (hi − lo) × min(h[lo], h[hi]). Start at the widest pair. Move the **shorter** side inward.

**Why:** say h[lo] ≤ h[hi]. Every container that uses lo with some other right end hi' < hi is narrower and its height is still capped by h[lo], so it can't beat the current one. Line lo is finished; drop it. Moving the taller side instead could skip the optimum.

## Trapping Rain Water

Water above i = min(maxLeft(i), maxRight(i)) − h[i]. Two pointers make it O(1) space: keep leftMax and rightMax. If leftMax < rightMax, the water at lo is decided: its right side is at least rightMax > leftMax, so the bound is leftMax. Add leftMax − h[lo] and move lo. Otherwise do the same on the right.

## Pattern 2: same direction (read and write pointers)

For in-place filtering, `write` marks where the next kept element goes, `read` scans everything:

```java
int write = 0;
for (int read = 0; read < n; read++) {
    if (keep(a[read])) a[write++] = a[read];
}
// a[0..write) is the answer
```

**Invariant:** a[0..write) always holds exactly the kept elements seen so far, in order. Remove Duplicates from Sorted Array, Move Zeroes and Remove Element are all this loop.

## Pattern 3: partitioning (Dutch national flag)

Sort 0s, 1s and 2s in one pass with three pointers: low (end of 0s), mid (scanner), high (start of 2s). Swap 0s to low, 2s to high; don't advance mid after swapping with high, because the swapped-in value hasn't been examined. This is also the partition step of quicksort and quickselect.

## Two pointers vs sliding window vs binary search

- Window: the answer is a **contiguous range** and validity is monotonic as it grows or shrinks.
- Two pointers (opposite ends): you look at **pairs**, and sorting gives a rule for which end to discard.
- Binary search: you look for one boundary; it's O(log n) per lookup, so "for each i, binary search a partner" is O(n log n), a fine fallback when the two-pointer argument isn't clear.

## Common bugs

- `while (lo <= hi)` when lo == hi would pair an element with itself.
- Forgetting duplicate skipping in 3Sum, or skipping before recording the first match.
- Sum overflow with large values: use `long`.

## Where it shows up

- **3Sum:** fix one, two pointers for the rest.
- **Container With Most Water:** move the shorter side.
- **Trapping Rain Water:** the side with the smaller max is decided.
- **Find the Duplicate Number, Reorder List, Remove Nth Node:** fast/slow pointers on lists.
- **Minimum Number of Operations to Make Array Continuous:** a same-direction window over sorted unique values.
