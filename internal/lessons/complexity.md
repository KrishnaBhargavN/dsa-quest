---
title: Reading constraints and complexity
group: Foundations
order: 1
topics:
minutes: 10
summary: The constraints tell you which algorithm the setter expects. Learn to read them before you think about code.
---

## Why this comes first

Your weak spot is choosing the algorithm under pressure. The single most reliable clue is sitting in the problem statement: the input size. A judge (or interviewer) allows roughly **10^8 simple operations per second**. Work backwards from n and you can rule out most approaches before you design anything.

## The table to memorise

| n up to | Budget | What it usually means |
| --- | --- | --- |
| 10–12 | O(n!) or O(n · 2^n) with a big constant | Permutations, brute force |
| 15–20 | O(2^n · n) | Bitmask DP, subsets |
| 100–500 | O(n^3) | Floyd-Warshall, interval DP, 3 nested loops |
| 1,000–5,000 | O(n^2) | 2D DP, all pairs, O(n^2) Prim |
| 10^5 – 10^6 | O(n log n) or O(n) | Sorting, heaps, binary search, two pointers, segment tree |
| 10^9 and above | O(log n) or O(sqrt n) | Binary search on the answer, math, fast power |

Interview problems often hide n. Ask for it. If the interviewer says "large", aim for O(n log n) or better.

## The diagnose-before-coding routine

Before writing a line, say out loud (or write in a comment):

1. **Tool:** which pattern? (window, heap, DP, BFS…)
2. **State:** what do I store? (a window's counts, dp[i][j], visited set)
3. **Transition:** how does one step change the state?
4. **Complexity:** does it fit the table above?

If step 4 fails, the tool is wrong. Go back to step 1, don't start optimising code.

## Signals in the wording

- "contiguous subarray / substring" → sliding window or prefix sums
- "sorted array", "minimum x such that…", "maximum of the minimum" → binary search
- "next greater / previous smaller" → monotonic stack
- "k largest / k closest / merge k" → heap
- "number of ways", "minimum cost", "can you reach" with overlapping choices → DP
- "all combinations / permutations / partitions" → backtracking
- "shortest path, unweighted" → BFS; "weighted, non-negative" → Dijkstra
- "dependencies / order of tasks" → topological sort
- "connected groups, merging" → union-find or DFS
- "prefix of words" → trie
- "range query with updates" → Fenwick or segment tree

## Amortised analysis, briefly

Some loops look quadratic but aren't. In a sliding window, `left` only moves forward, so across the whole run the inner `while` executes at most n times in total: O(n), not O(n^2). Same for a monotonic stack: each element is pushed once and popped at most once. When you see a nested loop, ask "how many times can the inner pointer move **in total**?"

## Java facts that change complexity

- `String` concatenation in a loop is O(n^2). Use `StringBuilder`.
- `ArrayList.remove(0)` is O(n). Use `ArrayDeque` for queues.
- `HashMap` is O(1) average; `TreeMap` is O(log n) but gives ordering (`floorKey`, `ceilingKey`).
- `Arrays.sort` on `int[]` is a dual-pivot quicksort; on objects it is a stable merge sort. Both O(n log n).
- Recursion depth beyond ~10^4 can overflow the stack. For 10^5 nodes in a line, go iterative.
- Overflow: `int` holds about 2.1 × 10^9. Sums of 10^5 values up to 10^9 need `long`. Products of two ints need `(long) a * b`.

## Space

Memory limits are usually 256 MB. An `int[10^7]` is 40 MB, fine. A `long[5000][5000]` is 200 MB, not fine; roll it to two rows.

## Checklist before you submit

- Did I check n = 0, n = 1, all equal values, negative numbers?
- Any `int` that can overflow?
- Any recursion that can go 10^5 deep?
- Any shared field that should have been a local variable? (Your recurring bug: state from one call leaking into the next. Prefer passing values as parameters or returning them.)
