---
title: Prefix sums and difference arrays
group: Foundations
order: 3
topics: Prefix sums and deque
minutes: 12
summary: Precompute running totals so any range sum is one subtraction, and turn range updates into two point updates.
---

## How to recognise it

- Many questions about sums over ranges [l, r].
- "Count subarrays whose sum is / is divisible by / is at least …".
- Many range updates ("add x to every index in [l, r]") followed by reads.
- Negative numbers present, so a sliding window can't decide when to shrink.

## The core idea

Let P[0] = 0 and P[i] = a[0] + … + a[i−1]. Then

sum(a[l..r]) = P[r+1] − P[l]

**Why:** P[r+1] is everything up to r; P[l] is everything before l. Subtracting removes exactly the elements before l. Using a length n+1 array with P[0] = 0 removes every special case for l = 0.

```java
long[] P = new long[n + 1];
for (int i = 0; i < n; i++) P[i + 1] = P[i] + a[i];
long rangeSum = P[r + 1] - P[l];
```

## Counting subarrays with a target sum

A subarray ending at j has sum k exactly when some earlier prefix equals P[j+1] − k. So walk left to right, and for each prefix ask "how many earlier prefixes equal current − k?" A map from prefix value to count answers that in O(1).

```java
Map<Long, Integer> count = new HashMap<>();
count.put(0L, 1);            // the empty prefix
long cur = 0; int ans = 0;
for (int x : a) {
    cur += x;
    ans += count.getOrDefault(cur - k, 0);
    count.merge(cur, 1, Integer::sum);
}
```

**Why the order matters:** you query before inserting the current prefix, so a subarray can't be empty.

Variants:
- **Divisible by k:** two prefixes give a sum divisible by k iff they are equal mod k. Count prefixes by `((cur % k) + k) % k`.
- **Longest subarray with sum k:** store the **first** index of each prefix value instead of the count.
- **Equal 0s and 1s:** map 0 → −1, then look for sum 0.

## 2D prefix sums

P[i][j] = sum of the rectangle from (0,0) to (i−1, j−1).

P[i+1][j+1] = a[i][j] + P[i][j+1] + P[i+1][j] − P[i][j]

Rectangle (r1,c1)–(r2,c2) = P[r2+1][c2+1] − P[r1][c2+1] − P[r2+1][c1] + P[r1][c1]. This is inclusion–exclusion: the top-left block is subtracted twice, so add it back once.

## Difference arrays: range updates in O(1)

To add v to every index in [l, r], do D[l] += v and D[r+1] −= v. After all updates, the prefix sum of D gives the final array.

**Why:** the prefix sum "turns on" +v at l and "turns it off" at r+1, so only indices in [l, r] receive it. It's the inverse of prefix sums: prefix sums answer range queries, difference arrays apply range updates. Fenwick trees do both online (see Range queries).

Car pooling, meeting-room counting and "Divide Intervals Into Minimum Number of Groups" are difference arrays in disguise: +1 at start, −1 after end, max prefix sum = max overlap.

## Prefix sums with a monotonic deque

When values can be negative and you want the **shortest** subarray with sum ≥ k, neither a map nor a plain window works. Keep a deque of prefix indices with **increasing** prefix values:

- From the front: while P[j] − P[front] ≥ k, record j − front and pop the front. It can never give a shorter answer later, because any later j is further away.
- From the back: while P[back] ≥ P[j], pop the back. A later start with a smaller or equal prefix is always at least as good: it's closer and leaves a larger sum.

Each index enters and leaves once: O(n). This is Shortest Subarray with Sum at Least K.

## Common bugs

- `int` prefix sums overflowing; use `long`.
- Forgetting the 0 prefix (or the {0: 1} seed).
- Negative modulo in Java: `-3 % 5 == -3`. Normalise with `((x % k) + k) % k`.
- Off-by-one between P indices and array indices. Stick to P of length n+1.

## Where it shows up

- **Subarray Sum Equals K**, **Path Sum III**: prefix count map.
- **Range Sum Query - Mutable**, **CSES Dynamic Range Sum**: prefix sums with updates need a Fenwick tree.
- **Range Update Queries (CSES)**: a difference array kept in a Fenwick tree.
- **Shortest Subarray with Sum at Least K**: prefix sums + monotonic deque.
