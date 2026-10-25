---
title: Recursion and backtracking
group: Trees and recursion
order: 10
topics: Backtracking
minutes: 15
summary: Build candidates one choice at a time, undo each choice after exploring it, and prune branches that can't succeed.
---

## How to recognise it

- "Return **all** subsets / combinations / permutations / partitions / placements."
- Small n (usually ≤ 15–20), because the output itself is exponential.
- Constraint satisfaction: N-Queens, Sudoku, word search in a grid.

If the problem asks for a **count** or an **optimum** rather than every solution, and states repeat, it's probably DP instead.

## Recursion first

Every recursive function needs: a **base case** that returns without recursing, and a **recursive case** that moves toward it. Trust the recursion: assume the call on a smaller input returns the right answer, and only check that you combine it correctly (this is induction).

## The backtracking template

```java
void backtrack(int start, List<Integer> path, List<List<Integer>> out) {
    if (isComplete(path)) {
        out.add(new ArrayList<>(path));        // copy! path keeps changing
        return;
    }
    for (int i = start; i < n; i++) {
        if (!canUse(i)) continue;              // prune
        path.add(nums[i]);                     // choose
        backtrack(i + 1, path, out);           // explore
        path.remove(path.size() - 1);          // un-choose
    }
}
```

**Why "un-choose" matters:** `path` is one shared list. After exploring everything that starts with a choice, you must restore the list exactly as it was so the next sibling choice starts clean. Forgetting it is the same bug as leaking state between calls. (You can avoid it entirely by passing a new list down, at the cost of copying.)

## The three families

**Subsets:** every element is in or out. Record the path at **every** node of the recursion tree (not just leaves), loop from `start`, and recurse with `i + 1`. 2^n subsets.

**Combinations with a target (Combination Sum):** loop from `start` so each combination is built in one canonical order (non-decreasing indices), which prevents duplicates like [2,3] and [3,2]. Reuse allowed: recurse with `i`; no reuse: `i + 1`. Sort first and `break` when the candidate exceeds the remaining target.

**Permutations:** order matters, so every unused element can go next. Track `boolean[] used`. n! results.

## Duplicates in the input

Sort, then at the same recursion depth skip an element equal to the previous one:

```java
if (i > start && nums[i] == nums[i - 1]) continue;
```

**Why `i > start` and not `i > 0`:** the rule is "don't start two sibling branches with the same value", because they'd produce identical results. Using the same value deeper in one branch (i == start) is fine and needed, e.g. [1, 1, 2]. For permutations with duplicates, the equivalent rule is: skip nums[i] if it equals nums[i − 1] and nums[i − 1] is not currently used.

## Grids: Word Search

DFS from each cell matching the first letter. Mark the cell as used (overwrite it with '#'), recurse into the four neighbours with the next index, then restore the letter. Return true as soon as any branch succeeds. Prune immediately on a mismatch.

## Constraint placement: N-Queens

One queen per row, so recurse over rows and choose a column. Track occupied columns and both diagonals in boolean arrays: cells on the same "\" diagonal share r − c (shift by n to keep it non-negative), cells on the same "/" diagonal share r + c. Each check is O(1).

## Partitioning: Palindrome Partitioning

Choose where the first piece ends; if s[start..end] is a palindrome, recurse on the rest. Precompute isPal[i][j] with a small DP (isPal[i][j] = s[i] == s[j] && isPal[i+1][j−1]) so each check is O(1).

## Complexity

Count leaves × work per leaf: subsets O(n · 2^n), permutations O(n · n!), combination sums depend on the target. Pruning doesn't change the worst case but matters a lot in practice.

## Common bugs

- Adding `path` itself instead of a copy to the output (everything ends up empty).
- Forgetting to undo the choice.
- Duplicate skipping with `i > 0` instead of `i > start`.
- Not restoring a grid cell after the DFS.

## Where it shows up

- **Subsets, Combination Sum I and II, Permutations, Word Search, Palindrome Partitioning, N-Queens.**
- **Word Search II:** backtracking guided by a trie.
- **Partition to K Equal Sum Subsets:** backtracking with pruning, or bitmask DP.
