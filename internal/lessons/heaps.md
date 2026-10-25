---
title: Heaps and priority queues
group: Trees and recursion
order: 12
topics: Heap
minutes: 13
summary: A heap gives you the smallest (or largest) item in O(log n) while items keep arriving. Use it for top-k, merging sorted streams, scheduling and running medians.
---

## How to recognise it

- "k largest / smallest / closest / most frequent".
- Repeatedly take the best available item while new items become available (scheduling, IPO, CPU tasks).
- Merge k sorted lists or streams.
- A running median.
- Dijkstra and Prim (see the graph lessons).

## Java's PriorityQueue

```java
PriorityQueue<Integer> minHeap = new PriorityQueue<>();
PriorityQueue<Integer> maxHeap = new PriorityQueue<>(Collections.reverseOrder());
PriorityQueue<int[]> byDist = new PriorityQueue<>((a, b) -> Integer.compare(a[0], b[0]));
```

`offer` and `poll` are O(log n), `peek` is O(1), `remove(Object)` is O(n): avoid it. Building from a collection is O(n).

## How a heap works (enough to reason about it)

A binary heap is an array-backed complete binary tree where each parent is ≤ its children. The minimum is at the root. Insert puts the item at the end and swaps it up while it's smaller than its parent; poll moves the last item to the root and swaps it down with its smaller child. Each is one root-to-leaf path: O(log n).

## Top k with a size-k heap

To keep the k **largest**, use a **min**-heap of size k: whenever it grows past k, poll the smallest. The heap always holds the k largest seen so far, and its root is the k-th largest.

**Why a min-heap for the largest:** the item you need to throw out is the weakest of your current top k, which is the minimum. O(n log k) time, O(k) space, and it works on streams.

Alternatives: quickselect (O(n) average, not streaming) and bucket sort when values are small counts (Top K Frequent).

## Two heaps: running median

A max-heap holds the lower half, a min-heap holds the upper half. Keep their sizes equal or the lower one bigger by one. Add each number to one side, then move the top across if the order is violated or the sizes are off by more than one. The median is the top of the lower half, or the average of both tops.

**Why it works:** the invariant "every number in the lower half ≤ every number in the upper half" plus balanced sizes means the tops are exactly the middle elements.

## Greedy with a heap

Many scheduling problems are: "advance time; make newly available items eligible; pick the best eligible item".

- **IPO:** sort projects by capital needed. While some are affordable, push their profits into a max-heap. Pick the top k times. Picking the most profitable affordable project is always right: it maximises capital, which can only unlock more projects.
- **Single-Threaded CPU:** sort tasks by enqueue time; a min-heap on (duration, index) holds the available ones. If the heap is empty, jump time forward to the next arrival.
- **Meeting Rooms III:** a heap of free rooms by number and a heap of busy rooms by (end time, number).
- **Task Scheduler:** a max-heap of counts simulates the schedule, though a counting formula is faster.

## k-way merge

Push the head of each sorted list; pop the smallest, then push the next element from the same list. Each of N elements is pushed and popped once in a heap of size k: O(N log k). Design Twitter's feed is a k-way merge of followees' tweet lists, stopping after 10.

## Lazy deletion

PriorityQueue can't efficiently delete an arbitrary element. Instead, leave stale entries and skip them when they reach the top (check against a "current value" array or a removed-count map). Dijkstra's `if (d > dist[u]) continue` is lazy deletion.

## Common bugs

- Comparators written as `a - b` overflow. Use `Integer.compare`.
- Min-heap vs max-heap mixed up for top-k.
- Calling `remove(obj)` in a loop: O(n) each.
- Mutating an object while it's inside the heap; the heap won't reorder it.

## Where it shows up

- **Top K Frequent Elements, K Closest Points, Task Scheduler, Single-Threaded CPU, Design Twitter, Find Median from Data Stream, IPO, Meeting Rooms III.**
- **Merge k Sorted Lists:** k-way merge.
- **Minimum Interval to Include Each Query:** sorted queries plus a heap of candidate intervals.
- **Dijkstra, Prim, Swim in Rising Water:** heaps in graph algorithms.
