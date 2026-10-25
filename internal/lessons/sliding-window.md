---
title: Sliding window
group: Foundations
order: 5
topics: Sliding window
minutes: 14
summary: Grow a contiguous window on the right and shrink it on the left; each index enters and leaves once, so it's O(n).
---

## How to recognise it

- "Longest / shortest / count of **contiguous** subarrays or substrings" with some condition.
- The condition is **monotonic**: if a window is valid, every smaller window inside it is valid (for "longest" problems), or every larger window containing it is valid (for "shortest" problems).
- All values non-negative (for sum conditions). Negatives break monotonicity; use prefix sums instead.

## The core idea

Brute force checks all O(n^2) windows. But if a window [l, r] is invalid for a "longest" problem, then [l, r'] for any r' > r is invalid too: it contains the bad window. So once r has made the window invalid, l must move. And l never needs to move back, because any window starting before l was already considered with a smaller-or-equal right end. Both pointers only move forward: at most 2n moves, so O(n).

## Template: longest valid window

```java
int left = 0, best = 0;
// state: counts, sums, distinct-count… for a[left..right]
for (int right = 0; right < n; right++) {
    add(a[right]);
    while (!valid()) {
        remove(a[left]);
        left++;
    }
    best = Math.max(best, right - left + 1);
}
```

**Invariant:** after the `while`, a[left..right] is the longest valid window ending at right. Taking the max over all right covers every possible ending point, so the answer is never missed.

## Template: shortest valid window

Flip the roles: shrink **while** the window is still valid, recording the answer inside the loop.

```java
for (int right = 0; right < n; right++) {
    add(a[right]);
    while (valid()) {
        best = Math.min(best, right - left + 1);
        remove(a[left]);
        left++;
    }
}
```

Minimum Window Substring is this shape. The state is `need[c]` (how many of c are still missing) and one integer `missing` (total characters still missing). `valid()` is `missing == 0`, an O(1) check.

## Fixed-size windows

For a window of exactly k, add a[right], and once right ≥ k remove a[right − k]. Permutation in String is a fixed window of size |s1| comparing 26 counts; keep a `matches` counter (how many letters have equal counts) to make each check O(1).

## Counting windows: "exactly k" = "at most k" − "at most k−1"

"Number of subarrays with exactly k distinct values" is not monotonic: adding an element can push you over k and you can't tell how far to shrink. "At most k" **is** monotonic, and with the longest-window template, every window ending at right with start ≥ left is valid, so it contributes `right − left + 1` subarrays. Then exactly(k) = atMost(k) − atMost(k − 1).

## The "replacement" trick

Longest Repeating Character Replacement: a window is valid if `length − maxFreq ≤ k`. You can keep `maxFreq` as the highest frequency ever seen, without decreasing it on shrink. **Why that's safe:** the answer only improves when a window with a higher maxFreq appears; a stale, too-high maxFreq can only make you keep a window at its previous best size, never report a bigger one wrongly. The window size never shrinks below the best found so far, which is fine because you only need the maximum.

## Window maximum or minimum: monotonic deque

To know the max of the current window in O(1), keep a deque of indices with **decreasing** values:

```java
Deque<Integer> dq = new ArrayDeque<>();
for (int i = 0; i < n; i++) {
    while (!dq.isEmpty() && a[dq.peekLast()] <= a[i]) dq.pollLast();
    dq.offerLast(i);
    if (dq.peekFirst() <= i - k) dq.pollFirst();   // fell out of the window
    if (i >= k - 1) out[i - k + 1] = a[dq.peekFirst()];
}
```

**Why popping smaller values is safe:** if a[j] ≤ a[i] and j < i, then j leaves the window before i does and is never larger, so j can never be a window maximum again. The deque holds only candidates that could still win, in decreasing order, so the front is the maximum.

With two deques (one for max, one for min) you get "longest subarray where max − min ≤ limit".

## Common bugs

- Using a window when values can be negative (sum conditions stop being monotonic).
- Recording the answer at the wrong moment (inside vs after the shrink loop).
- `while` vs `if` for shrinking: a variable window needs `while`.
- `HashMap` counts that reach 0 but stay in the map, breaking a `map.size()` distinct count. Remove keys at 0.

## Where it shows up

- **Longest Substring Without Repeating Characters:** last-seen index lets left jump directly.
- **Longest Repeating Character Replacement:** length − maxFreq ≤ k.
- **Permutation in String:** fixed window, 26 counts.
- **Max Consecutive Ones III:** at most k zeros.
- **Minimum Window Substring:** shortest-window template.
- **Sliding Window Maximum:** monotonic deque.
- **Subarrays with K Different Integers:** atMost(k) − atMost(k−1).
- **Longest Continuous Subarray With Absolute Diff ≤ Limit:** two deques.
