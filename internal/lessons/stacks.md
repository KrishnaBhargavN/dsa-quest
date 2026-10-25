---
title: Stacks and monotonic stacks
group: Foundations
order: 8
topics: Stack and monotonic stack
minutes: 15
summary: A stack handles "most recent unfinished thing"; a monotonic stack finds the next or previous greater/smaller element for every index in O(n).
---

## How to recognise it

- Nested structure: brackets, `3[a2[c]]`, expressions, undo, directory paths.
- "Next greater element", "previous smaller", "how many days until warmer", "span".
- "Largest rectangle", "sum of subarray minimums": each element's reach to the left and right.
- Removing characters to make the result lexicographically smallest.

## Plain stacks: the most recent open thing

A stack is right whenever the thing you need to resolve next is the **most recent** unresolved one. Brackets: a closing bracket must match the most recent unmatched opening one. Expressions in postfix: an operator applies to the two most recent values.

Decode String keeps two stacks: counts and partial strings. On `[`, push the current string and count and start fresh. On `]`, pop and append the current string k times to the popped one. Nesting is handled automatically because the stack restores the outer context.

In Java use `ArrayDeque` (`push`, `pop`, `peek`), not `Stack` (synchronised and slower).

## Monotonic stack: the core idea

To find, for every i, the next index j > i with a[j] > a[i]:

```java
int[] next = new int[n];
Arrays.fill(next, -1);
Deque<Integer> st = new ArrayDeque<>();     // indices, values decreasing from bottom to top
for (int i = 0; i < n; i++) {
    while (!st.isEmpty() && a[st.peek()] < a[i]) {
        next[st.pop()] = i;                  // i is the first greater element for that index
    }
    st.push(i);
}
```

**Why it's correct:** the stack holds exactly the indices still waiting for a greater element. When a[i] arrives, every waiting index with a smaller value has found its answer, and i is the **first** greater one, because if an earlier index k had been greater, it would have popped them already. Indices with values ≥ a[i] keep waiting, and they sit below i, so the stack stays decreasing.

**Why it's O(n):** each index is pushed once and popped at most once.

The four variants:
- Next greater: scan left to right, pop while top < current (stack decreasing).
- Next smaller: pop while top > current (stack increasing).
- Previous greater or smaller: the element **below** an index on the stack when it's pushed is its previous greater or smaller.
- Strict vs non-strict comparison decides how ties are handled. Choose deliberately.

## Contribution technique (Sum of Subarray Minimums)

Instead of enumerating subarrays, ask how many subarrays have a[i] as their minimum. If L is how far you can extend left before hitting a smaller element, and R how far right, then a[i] is the minimum of L × R subarrays. Total = Σ a[i] × L × R.

**Ties:** with equal values, a subarray would be counted for both. Use "strictly smaller" on one side and "smaller or equal" on the other so each subarray is attributed to exactly one minimum (say, the leftmost).

## Largest Rectangle in Histogram

The best rectangle that uses bar i at full height extends left and right until a shorter bar. With an increasing stack, when bar i pops index top (because h[i] < h[top]), you know both ends at once: the right boundary is i, and the left boundary is the new top of the stack. Width = i − newTop − 1 (or i if the stack is empty). Push a final height 0 to flush everything.

Maximal Rectangle builds a histogram per row (heights of consecutive 1s) and runs this on each row.

## Greedy with a stack: smallest number after removals

Remove K Digits: scanning left to right, if the previous kept digit is larger than the current one and you still have removals, remove it. A smaller digit earlier always beats anything later in the number, so removing a larger digit before a smaller one is always right. That's an increasing stack. Leftover removals come off the end; strip leading zeros.

## Car Fleet: a stack of arrival times

Sort cars by position from closest to the target. A car that would arrive no later than the fleet in front of it catches up and joins it. Keep a stack (or just the current fleet's time): a new fleet forms only when a car's time is strictly greater.

## Common bugs

- Pushing values instead of indices when you need widths or distances.
- Wrong tie handling, so you double count or miss subarrays.
- Forgetting to flush the stack at the end (append a sentinel).
- `Integer` comparison with `==` on values from a `Deque<Integer>`.

## Where it shows up

- **Evaluate Reverse Polish Notation, Decode String:** plain stacks.
- **Daily Temperatures:** next greater.
- **Car Fleet:** stack of arrival times.
- **Remove K Digits:** greedy increasing stack.
- **Sum of Subarray Minimums:** contribution with previous/next smaller.
- **Largest Rectangle in Histogram, Maximal Rectangle.**
- Monotonic **deques** (a stack you can also pop from the front) are in the Sliding window and Prefix sums lessons.
