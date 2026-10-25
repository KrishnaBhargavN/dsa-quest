---
title: Grid DP
group: Dynamic programming
order: 23
topics: DP - grid
minutes: 10
summary: When moves only go right and down, each cell depends on cells above and to the left, so you can fill the grid row by row.
---

## How to recognise it

- A grid or triangle, and moves restricted to one direction (right/down, or down-left/down/down-right).
- "Number of paths", "minimum path sum", "largest square".
- If moves can go in every direction, it's a graph problem (BFS/Dijkstra), not grid DP, because cells would depend on each other in a cycle.

## Why restricted moves make it a DP

With only right and down moves, any path to (i, j) arrives from (i−1, j) or (i, j−1). Both are "earlier" in row-major order, so filling rows top to bottom and columns left to right always computes dependencies first. No cycles, so no need for Dijkstra.

## Templates

Count paths:

```java
int[] dp = new int[n];
dp[0] = 1;
for (int i = 0; i < m; i++)
    for (int j = 0; j < n; j++) {
        if (grid[i][j] == 1) dp[j] = 0;          // obstacle
        else if (j > 0) dp[j] += dp[j - 1];      // dp[j] still holds the value from above
    }
return dp[n - 1];
```

**Why one row works:** before the update, dp[j] holds row i−1's value (from above), and dp[j−1] was already updated to row i (from the left). Adding them gives exactly paths from above plus paths from the left.

Minimum path sum is the same with `min` and the cell cost: dp[j] = grid[i][j] + min(dp[j], dp[j−1]), treating out-of-bounds as infinity, and the first row and column as running sums.

## Go backwards when the future constrains the past

**Triangle:** top-down, the edges of each row need special cases. Bottom-up from the last row, dp[j] = row[j] + min(dp[j], dp[j+1]) has no edge cases and the answer ends up in dp[0].

**Dungeon Game:** going forward, the minimum health needed depends on both your current health and the lowest it has dipped, which is two numbers. Going backward, the state is one number: need[i][j] = the minimum health on entering (i, j) to survive the rest. need[i][j] = max(1, min(need of right, need of down) − dungeon[i][j]). The max(1, …) says you must always have at least 1 health.

**General lesson:** when the forward state needs to carry history, try defining the state as "what's needed from here to the end".

## Shapes beyond paths: Maximal Square

dp[i][j] = side of the largest all-1 square with bottom-right corner (i, j) = 1 + min(top, left, top-left) when the cell is 1.

**Why the min of three:** a square of side s at (i, j) needs squares of side s − 1 ending at the cell above, the cell to the left, and the cell diagonally up-left (they overlap to cover everything except the corner). The smallest of the three limits the size.

## Falling paths

Minimum Falling Path Sum: each cell in row r depends on columns c−1, c, c+1 of row r−1. Same row-by-row fill, three candidates instead of two. Clamp at the edges.

## Common bugs

- Initialising the first row and column incorrectly (an obstacle in row 0 blocks everything after it in that row).
- Overwriting values you still need when using one row (loop direction matters; check which neighbours are "old" and which "new").
- Using DP when moves go in all four directions.

## Where it shows up

- **Unique Paths I and II, Minimum Path Sum, Triangle, Minimum Falling Path Sum, Maximal Square, Dungeon Game.**
- **Maximal Rectangle** looks like grid DP but is solved with histograms and a monotonic stack.
- String DPs (LCS, Edit Distance) are grid DPs where the grid is "prefix of s × prefix of t".
