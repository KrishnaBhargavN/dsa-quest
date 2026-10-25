---
title: Dynamic programming: how to design one
group: Dynamic programming
order: 22
topics: DP - linear
minutes: 18
summary: DP is recursion that never solves the same subproblem twice. Define the state, write the transition, fix the base cases, then choose the order.
---

## How to recognise it

- "Number of ways", "minimum cost", "maximum value", "is it possible" over a sequence of choices.
- A brute force recursion that makes a choice at each step and **revisits the same situation** through different paths.
- Greedy feels tempting but you can build a counterexample.

## Why DP works: two properties

1. **Optimal substructure:** the best answer for a situation is built from best answers to smaller situations. (The best way to rob houses 0..i uses the best way to rob 0..i−1 or 0..i−2.)
2. **Overlapping subproblems:** the same smaller situation is reached many ways. Plain recursion recomputes it exponentially often; storing it once makes the total work (number of states) × (work per state).

If there's no overlap, plain recursion or divide and conquer is enough. If there's no optimal substructure, DP gives wrong answers.

## The four-step recipe

1. **State:** what is the smallest description of "where I am" that determines the best future? Write it in words: "dp[i] = max money robbing only houses 0..i".
2. **Transition:** the last choice made. "At house i, either rob it (a[i] + dp[i−2]) or skip it (dp[i−1])."
3. **Base cases:** the smallest states, answered directly.
4. **Order:** compute each state after the states it depends on.

If you can't write the state in one sentence, you don't have it yet. Most wrong DPs are a state that forgets something the future depends on.

## Top-down vs bottom-up

```java
// Top-down: recursion + memo. Natural to write; computes only reachable states.
Integer[] memo;
int best(int i) {
    if (i < 0) return 0;
    if (memo[i] != null) return memo[i];
    return memo[i] = Math.max(best(i - 1), a[i] + best(i - 2));
}

// Bottom-up: a loop in dependency order. No recursion depth limit; easy to save space.
int prev2 = 0, prev1 = 0;
for (int x : a) {
    int cur = Math.max(prev1, x + prev2);
    prev2 = prev1; prev1 = cur;
}
```

Write top-down first if the state is unclear; convert to bottom-up when you need space savings or n is large enough for stack overflow. Keep the memo as a parameter or a field you reset per call, never a static that leaks between test cases.

## Space optimisation

If dp[i] only reads dp[i−1] and dp[i−2], keep two variables. If a 2D dp[i][j] only reads row i−1, keep two rows (or one row with the right loop direction; see the Knapsack lesson).

## Linear DP patterns

- **Take or skip (House Robber):** dp[i] = max(dp[i−1], a[i] + dp[i−2]). House Robber II on a circle: the first and last can't both be taken, so solve [0..n−2] and [1..n−1] separately and take the max. That covers every valid choice because any valid plan excludes the first or the last house.
- **Reduce to a known DP (Delete and Earn):** taking value v forbids v−1 and v+1, so total[v] = v × count(v) turns it into House Robber over values.
- **Count ways (Decode Ways, Climbing Stairs):** dp[i] = sum over the last step's options. Decode Ways: one digit (if not '0') plus two digits (if 10..26).
- **Segmentation (Word Break):** dp[i] = some j < i with dp[j] and s[j..i) a word. The state is "this prefix is splittable".
- **Max/min product (Maximum Product Subarray):** a negative flips the order, so keep both the max and the min ending at i. This is "the state must remember what the future needs": the future needs the min too.
- **Ending-at-i (Kadane):** best subarray ending at i = max(a[i], a[i] + best ending at i−1). The answer is the max over all i.

## Longest Increasing Subsequence

O(n^2): dp[i] = 1 + max dp[j] for j < i with a[j] < a[i].

O(n log n): tails[k] = the smallest possible last value of an increasing subsequence of length k + 1. For each x, find the first tail ≥ x (lower bound) and replace it with x; if none, append.

**Why it's correct:** tails stays strictly increasing (a longer subsequence's tail can't be smaller than a shorter one's best tail). Replacing a tail with a smaller value never hurts: it keeps the same length with more room to extend. The length of tails is the LIS length. (tails itself isn't an actual subsequence; only its length is meaningful.)

## When it's greedy, not DP (Jump Game II)

Jump Game II looks like DP (dp[i] = min jumps to i) but has a greedy O(n) solution: indices reachable in exactly j jumps form a contiguous range, so you can BFS by ranges. Before writing a DP, ask whether the choices have a structure (ranges, sortedness) that makes one choice always best. If you can prove it with an exchange argument, greedy is simpler.

## Common bugs

- A state that forgets necessary information (e.g. only the max in Maximum Product Subarray).
- Wrong base case for empty input (dp[0] = 1 for "number of ways", 0 for "max value").
- Using `int` when counts can exceed 2^31; many problems ask you to mod 1e9+7.
- A memo stored in a static field that isn't reset between calls.

## Where it shows up

- **House Robber II, Delete and Earn, Decode Ways, Word Break, Maximum Product Subarray.**
- **Longest Increasing Subsequence:** redo it cold in O(n log n).
- **Jump Game II:** check whether greedy suffices first.
- **Jump Game VI:** dp[i] = a[i] + max of the last k dp values, sped up with a monotonic deque.
- Every other DP lesson builds on this one.
