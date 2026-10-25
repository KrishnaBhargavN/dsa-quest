---
title: Sorting, quickselect and counting tricks
group: Foundations
order: 7
topics:
minutes: 12
summary: Sorting is often the whole first step. Know when sorting is enough, when O(n) selection beats it, and the comparator pitfalls in Java.
---

## How to recognise it

- Order doesn't matter in the input but does in the logic ("pairs", "intervals", "closest").
- "k-th smallest / largest" or "top k".
- Counting inversions or "smaller elements after self".
- Values in a small range (counting sort, bucket sort).

## Sorting as preprocessing

Sorting costs O(n log n) and often unlocks O(n) logic afterwards: two pointers, greedy sweeps, interval merging, grouping equal values. If your brute force is O(n^2) and order doesn't matter, ask "what if it were sorted?"

## Java comparators

```java
Arrays.sort(intervals, (a, b) -> Integer.compare(a[0], b[0]));
list.sort(Comparator.comparingInt((int[] x) -> x[1]).thenComparingInt(x -> x[0]));
```

Never write `(a, b) -> a[0] - b[0]`: subtraction overflows for large or negative values and silently gives the wrong order. Use `Integer.compare`.

To sort an `int[]` descending, box it (`Integer[]`) or sort ascending and read backwards.

## Merge sort and counting inversions

Merge sort splits in half, sorts each half, and merges. While merging, when you take an element from the right half, every element still waiting in the left half is larger and came before it: that's an inversion. Summing those counts during merges counts all inversions in O(n log n).

**Why it counts every pair once:** any pair (i < j) is split between the two halves at exactly one level of recursion, the level where they stop being in the same subarray, and it is counted at that merge.

Count of Smaller Numbers After Self uses the same idea (track original indices), or a Fenwick tree over compressed values.

## Quickselect: k-th element in O(n) average

Partition around a pivot like quicksort, then recurse into **only** the side containing index k.

```java
int quickselect(int[] a, int k) {          // k-th smallest, 0-based
    int lo = 0, hi = a.length - 1;
    Random rnd = new Random();
    while (true) {
        int p = lo + rnd.nextInt(hi - lo + 1);
        swap(a, p, hi);
        int store = lo;
        for (int i = lo; i < hi; i++) if (a[i] < a[hi]) swap(a, i, store++);
        swap(a, store, hi);                 // pivot is now at its final position
        if (store == k) return a[store];
        if (store < k) lo = store + 1; else hi = store - 1;
    }
}
```

**Why it's O(n) on average:** a random pivot lands in the middle half with probability 1/2, shrinking the range to at most 3/4. The work is n + (3/4)n + (9/16)n + … = O(n). The worst case is O(n^2), which a random pivot makes vanishingly unlikely.

Compared with a heap of size k (O(n log k)), quickselect is faster but mutates the array and isn't streaming.

## Counting sort and bucket sort

When values fit in a small range (letters, ages, frequencies ≤ n), count occurrences in an array: O(n + range). Top K Frequent Elements in O(n): bucket i holds the values with frequency i; read buckets from high to low.

## Cyclic sort (values 1..n)

When an array holds values in 1..n, each value has a "home" index value − 1. Swap each element home until the current slot is right. Every swap puts one value home for good, so it's O(n). Use it for "first missing positive", "find all duplicates", "find the missing number" in O(1) extra space.

## Coordinate compression

When values are huge (10^9) but there are few of them (10^5), sort the distinct values and replace each value with its rank. Now you can index arrays and Fenwick trees by rank. Order is preserved, which is all most algorithms need.

## Stability

A stable sort keeps equal elements in their original order. Java's object sort is stable; that lets you sort by a secondary key first and the primary key second.

## Common bugs

- Subtraction comparators overflowing.
- Sorting the array when the original indices are part of the answer; sort an index array instead (`Integer[] idx` with a comparator on values).
- Forgetting quickselect mutates the input.

## Where it shows up

- **Merge Intervals, Non-overlapping Intervals, Meeting Rooms:** sort then sweep.
- **K Closest Points to Origin:** quickselect or a size-k heap.
- **Top K Frequent Elements:** bucket sort.
- **Count of Smaller Numbers After Self:** merge sort counting or Fenwick + compression.
- **Find the Duplicate Number:** the cyclic-sort idea is ruled out (no modifying), which is why it becomes a cycle-detection problem.
