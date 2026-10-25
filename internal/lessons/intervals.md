---
title: Intervals and sweep lines
group: Greedy and intervals
order: 15
topics: Intervals
minutes: 12
summary: Sort by start (to merge) or by end (to keep as many as possible), or turn intervals into +1/−1 events and sweep to count overlaps.
---

## How to recognise it

- Input is pairs [start, end]: meetings, bookings, ranges.
- "Merge overlapping", "insert", "minimum removals", "minimum rooms", "maximum overlap".
- Queries asking which interval covers a point.

First decide: are endpoints inclusive? [1, 2] and [2, 3]: do they overlap? The problem statement decides, and it changes `<` vs `<=` everywhere.

## Merge: sort by start

```java
Arrays.sort(iv, (a, b) -> Integer.compare(a[0], b[0]));
List<int[]> out = new ArrayList<>();
for (int[] cur : iv) {
    if (out.isEmpty() || out.get(out.size() - 1)[1] < cur[0]) out.add(cur);
    else out.get(out.size() - 1)[1] = Math.max(out.get(out.size() - 1)[1], cur[1]);
}
```

**Why sorting by start is enough:** after sorting, an interval can only overlap the **last** merged block. Any earlier block ends before the last one starts, and the current interval starts after the last one starts, so it can't reach back.

Insert Interval doesn't even need sorting (the input is sorted): copy intervals that end before the new one, absorb every interval that starts no later than the new end, copy the rest.

## Keep the most / remove the fewest: sort by end

Non-overlapping Intervals: keep an interval if it starts at or after the last kept end; answer = n − kept. Sorting by **end** is the key: the interval that ends first leaves the most room for the rest (exchange argument in the Greedy lesson). Sorting by start fails: one long early interval would block many short ones.

## Counting overlaps: sweep line

Turn each interval into two events: +1 at start, −1 just after end. Sort events (at equal times, process −1 before +1 if touching intervals don't overlap). Sweep, keeping a running count; the maximum is the peak overlap.

**Why the peak overlap is the minimum number of groups (rooms):** you need at least that many, since those intervals pairwise overlap at that moment. And it's enough: process intervals by start and reuse any group whose last interval has ended. A new group is only opened when all current groups are busy, which happens only while the overlap equals the current number of groups.

With a heap instead of events: sort by start; a min-heap holds the end times of active groups; if the earliest end is before the new start, reuse it (poll), then push the new end. The heap size is the answer.

## Point queries: sort both and sweep

Minimum Interval to Include Each Query: sort intervals by start and queries by value (remember original indices). For each query q in increasing order:
1. Push every interval with start ≤ q into a min-heap keyed by (size, end).
2. Pop intervals whose end < q; they can't contain this or any later query.
3. The top is the smallest interval containing q.

Each interval is pushed and popped once: O((n + q) log n). Answering queries **offline** (in an order you choose) is a powerful general trick.

## Interval DP is different

"Cut a stick", "burst balloons", "merge stones" are not interval-list problems; they're DP over ranges [i, j]. See the Interval DP lesson.

## Common bugs

- Inclusive vs exclusive endpoints.
- Comparator overflow with `a[0] - b[0]` on large or negative values.
- Mutating input intervals that are still referenced elsewhere.
- Sorting by start for a maximise-kept problem.

## Where it shows up

- **Merge Intervals, Insert Interval, Non-overlapping Intervals, Divide Intervals Into Minimum Number of Groups, Minimum Interval to Include Each Query.**
- **Meeting Rooms III:** interval simulation with two heaps.
- **Car Pooling, My Calendar:** sweep lines and difference arrays.
