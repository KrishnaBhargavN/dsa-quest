---
title: Bitmask DP
group: Dynamic programming
order: 29
topics: DP - bitmask
minutes: 14
summary: When n ≤ ~20, a set of chosen items fits in an int. DP over all 2^n subsets turns factorial searches into O(2^n · n).
---

## How to recognise it

- n is small: ≤ 12 to 20. This is the loudest signal (see the constraints lesson).
- You need to remember **which** items are used, not just how many: assignments, visiting every node, packing items into bins, partitioning into groups.
- A brute force over permutations would be n! (20! ≈ 2.4 × 10^18), but over subsets it's 2^n (2^20 ≈ 10^6).

## Bit operations you need

```java
(mask >> i) & 1          // is item i in the set?
mask | (1 << i)          // add item i
mask & ~(1 << i)         // remove item i
Integer.bitCount(mask)   // size of the set
(1 << n) - 1             // the full set
for (int sub = mask; sub > 0; sub = (sub - 1) & mask) { }   // all non-empty subsets of mask
```

## The core idea

Permutation search wastes time because many orders lead to the same situation: having used items {0, 2, 5} is the same whether you used them as 0, 2, 5 or 5, 0, 2. If the future only depends on **which** items are used (and maybe one more small piece, like the last item), make that the state.

## Travelling-salesman shape: (mask, last)

dp[mask][v] = the best cost to have visited exactly the nodes in mask, ending at v. Transition: go to an unvisited u. O(2^n · n^2).

**Shortest Path Visiting All Nodes** is this with unweighted edges, so BFS over (node, mask) states replaces the DP. Start from every node at once (multi-source), because any start is allowed. The first time a state with the full mask is popped, its distance is the answer.

## Packing shape: (mask) → (bins used, fill of the current bin)

Work Sessions and Elevator Rides: put items into bins of capacity C, minimising the number of bins. The state is the set of packed items; the value is the pair (bins used, space used in the last open bin), compared lexicographically.

```java
int[][] best = new int[1 << n][];               // {bins, lastFill}
best[0] = new int[]{1, 0};
for (int mask = 0; mask < (1 << n); mask++) {
    if (best[mask] == null) continue;
    for (int i = 0; i < n; i++) {
        if ((mask >> i & 1) != 0) continue;
        int[] cur = best[mask];
        int[] next = cur[1] + w[i] <= C
            ? new int[]{cur[0], cur[1] + w[i]}
            : new int[]{cur[0] + 1, w[i]};
        int nm = mask | (1 << i);
        if (best[nm] == null || compare(next, best[nm]) < 0) best[nm] = next;
    }
}
// answer: best[(1 << n) - 1][0]
```

**Why lexicographic (bins, fill) is a valid state value:** for the same set of packed items, fewer bins is always better, and with equal bins, less fill in the open bin leaves more room. One state dominates another exactly when it's smaller in this order, so keeping only the best is safe. O(2^n · n). For CSES Elevator Rides (n = 20) use two flat `int[]` arrays instead of `int[][]` for speed.

## Partition into k equal groups

Partition to K Equal Sum Subsets is the same packing idea with capacity = total / k, where every bin must be exactly full. dp[mask] = fill of the current bin (mod target), or −1 if unreachable; only add item i if it fits in the current bin.

## Subset enumeration: O(3^n)

Iterating all submasks of every mask totals 3^n (each item is in the submask, in the mask but not the submask, or outside). Fine for n ≤ 15 or so. Used for "split into groups, each with a cost" problems.

## Common bugs

- Operator precedence: `mask >> i & 1` is `(mask >> i) & 1` in Java, but write the parentheses to be safe; `1 << n - 1` is `1 << (n - 1)`.
- Memory: dp[1 << 20][20] of ints is 80 MB. Check before allocating.
- Iterating masks in the wrong order: adding items only increases the mask, so increasing numeric order is a valid order.

## Where it shows up

- **Partition to K Equal Sum Subsets, Minimum Number of Work Sessions, Shortest Path Visiting All Nodes, Elevator Rides (CSES).**
- Assignment problems (n workers, n jobs) with n ≤ 20.
