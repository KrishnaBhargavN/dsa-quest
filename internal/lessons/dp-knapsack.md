---
title: Knapsack DP
group: Dynamic programming
order: 24
topics: DP - knapsack
minutes: 15
summary: Pick items under a capacity. 0/1 knapsack loops capacity downward, unbounded loops upward, and many "subset" problems reduce to one of them.
---

## How to recognise it

- Items, each with a cost (weight, price, size), and a capacity or target.
- "Can you reach exactly X?", "how many ways to make X?", "max value within budget X?", "fewest items to make X?"
- Partitions into two groups with some balance: these reduce to "subset with sum S".
- The target is small enough to index an array (up to ~10^5–10^6).

## 0/1 knapsack: each item at most once

State: dp[c] = best value using the items seen so far with total weight ≤ c (or exactly c).

```java
int[] dp = new int[cap + 1];
for (int i = 0; i < n; i++)
    for (int c = cap; c >= w[i]; c--)            // downward!
        dp[c] = Math.max(dp[c], dp[c - w[i]] + v[i]);
```

**Why downward:** dp[c − w[i]] must still mean "before item i". Going down, you read smaller capacities that haven't been updated for item i yet, so each item is used at most once. Going up would let item i build on itself.

The full 2D form is dp[i][c] = max(dp[i−1][c], dp[i−1][c − w] + v): skip item i, or take it. The 1D array is that table with rows overwritten in place.

## Unbounded knapsack: reuse allowed

Same loop with capacity going **upward**: now dp[c − w] may already include item i, which is exactly "use it again".

**Coin Change (fewest coins):** dp[a] = min over coins of dp[a − coin] + 1, dp[0] = 0, others start at infinity (amount + 1 works as a safe infinity).

## Counting: combinations vs permutations

```java
// Combinations (order doesn't matter): coins outer, amounts inner
dp[0] = 1;
for (int coin : coins) for (int a = coin; a <= amount; a++) dp[a] += dp[a - coin];

// Permutations (order matters): amounts outer, coins inner
dp[0] = 1;
for (int a = 1; a <= amount; a++) for (int coin : coins) if (a >= coin) dp[a] += dp[a - coin];
```

**Why the loop order decides it:** with coins outer, by the time you use coin k, all uses of coins before k are already counted, so each multiset is built in one fixed coin order: counted once. With amounts outer, every coin can be the "last" one at every amount, so 1+2 and 2+1 are both counted. Coin Change II wants combinations.

## Reductions to subset sum

- **Partition Equal Subset Sum:** two equal halves exist iff some subset sums to total / 2 (odd total → impossible). Boolean 0/1 knapsack. A `BitSet` shifted left by each number is a fast trick: `bits |= bits << x`.
- **Target Sum:** with P = plus set, sum(P) − (total − sum(P)) = target, so sum(P) = (total + target) / 2. Count subsets with that sum; zero ways if it's negative or odd.
- **Last Stone Weight II:** every smash sequence ends as a signed sum of the stones, so the result is |sumA − sumB| for some split. Find the reachable subset sum closest to total / 2 from below: answer total − 2s.

## Two capacities: Ones and Zeroes

Each string costs (zeros, ones). dp[i][j] = the most strings using at most i zeros and j ones. Same 0/1 rule: loop both i and j downward per string.

## Fast I/O for CSES in Java

CSES time limits are tight for Java. Read input with a `DataInputStream`-based reader or `StreamTokenizer`, never `Scanner`, and print with a `StringBuilder` or `PrintWriter`. Book Shop (n = 1000, x = 10^5) is 10^8 simple steps: the 1D downward loop with an `int[]` passes; a 2D `int[1001][100001]` (400 MB) doesn't.

## Common bugs

- Wrong loop direction (0/1 vs unbounded).
- Initialising "impossible" states as 0 in a min-count problem.
- Overflow when counting ways; check whether a modulus is required.
- Forgetting the parity check in Target Sum and Partition.

## Where it shows up

- **Coin Change, Coin Change II, Partition Equal Subset Sum, Target Sum, Ones and Zeroes, Last Stone Weight II, Book Shop (CSES).**
- **Word Break** is an unbounded "can I fill the length?" DP in disguise.
