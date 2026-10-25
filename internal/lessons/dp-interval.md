---
title: Interval DP
group: Dynamic programming
order: 26
topics: DP - interval
minutes: 13
summary: dp[i][j] is the best answer for the range i..j, built by choosing a split point k. Fill by increasing range length. The hard part is choosing what k means.
---

## How to recognise it

- A sequence where you repeatedly **merge, cut, burst or remove** pieces, and the cost depends on neighbours or the current piece.
- Constraints like n ≤ 100–500 (the solution is O(n^3)).
- "Optimal order of operations" on a line or a polygon.
- Games where two players take from either end (minimax over ranges).

## The shape

```java
for (int len = 2; len <= n; len++)              // range length
    for (int i = 0; i + len - 1 < n; i++) {
        int j = i + len - 1;
        dp[i][j] = INF;
        for (int k = i; k < j; k++)              // split point
            dp[i][j] = Math.min(dp[i][j], dp[i][k] + dp[k + 1][j] + cost(i, k, j));
    }
```

**Why increasing length:** dp[i][j] only reads strictly shorter ranges, so they're all final by the time you need them.

## Choosing what k means: first or last?

This is the whole difficulty. The split must make the two sides **independent**.

**Cut a stick: k is the first cut.** Sort the cut positions and add 0 and n as endpoints c[0..m+1]. dp[i][j] = cheapest way to make every cut strictly between c[i] and c[j]. If the first cut in that segment is at c[k], it costs the segment length c[j] − c[i], and splits it into two segments that never interact again: dp[i][k] + dp[k][j] + (c[j] − c[i]).

**Burst Balloons: k is the last balloon.** If you pick the **first** balloon to burst, its neighbours change and the two sides become entangled. Instead pick the balloon k that is burst **last** within the open range (i, j). While the rest of the range is being burst, k is still there, so the two sides (i, k) and (k, j) are separated by k and are independent. When k finally bursts, its neighbours are i and j (the range boundaries): dp[i][j] = max over k of dp[i][k] + dp[k][j] + a[i]·a[k]·a[j]. Pad the array with 1 on both ends.

**Polygon triangulation: k is the third vertex.** Edge (i, j) belongs to exactly one triangle (i, k, j) with i < k < j. That triangle splits the polygon into two smaller polygons: dp[i][k] + dp[k][j] + v[i]·v[k]·v[j].

The question to ask yourself: "Which choice, once made, separates the problem into two parts that don't affect each other?"

## Minimax over ranges

Guess Number Higher or Lower II: guessing k in [l, r] costs k plus the worse of the two sides, because the adversary picks the side: dp[l][r] = min over k of k + max(dp[l][k−1], dp[k+1][r]).

Two-player "take from either end" games (Stone Game, Predict the Winner): dp[i][j] = the best score difference the current player can achieve on [i, j] = max(a[i] − dp[i+1][j], a[j] − dp[i][j−1]). The minus sign is because the opponent then plays optimally on what's left.

## Complexity

O(n^2) states × O(n) splits = O(n^3). n = 500 gives about 2 × 10^7 inner steps, fine.

## Common bugs

- Picking "first" when the sides interact (Burst Balloons). If the cost of a choice depends on things that change later, try "last".
- Filling by i then j (rows) instead of by length, so you read unfinished values. (Alternatively, i from n − 1 down and j from i + 1 up also works.)
- Off-by-one between inclusive and exclusive ranges; decide once and write it in a comment.

## Where it shows up

- **Minimum Score Triangulation, Guess Number Higher or Lower II, Minimum Cost to Cut a Stick, Burst Balloons.**
- **Longest Palindromic Subsequence** is interval DP over a string.
