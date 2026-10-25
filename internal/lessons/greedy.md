---
title: Greedy algorithms and how to prove them
group: Greedy and intervals
order: 14
topics: Greedy
minutes: 15
summary: Make the locally best choice and never look back. It's only correct when you can prove no better solution was given up; the exchange argument is how.
---

## How to recognise it

- A choice at each step, and one option looks obviously best ("take the earliest finishing", "go as far as possible").
- Sorting by some key seems to make the problem easy.
- Constraints too big for DP (n = 10^5 with a large value range).

The danger: greedy ideas feel right and are often wrong. Always try to break your greedy with a small counterexample before coding. If you can't break it, prove it.

## Proof technique 1: the exchange argument

Take any optimal solution that differs from the greedy one. Find the first place they differ, and show you can **swap** the optimal's choice for the greedy choice without making it worse. Repeating the swap turns the optimal solution into the greedy one, so the greedy one is optimal too.

Example, interval scheduling (keep the most non-overlapping intervals): greedy picks the interval that **ends first**. Suppose an optimal solution picks a different first interval X. The greedy choice G ends no later than X, so replacing X with G can't overlap anything that came after X. Same count, now agreeing with greedy at step one. Induct.

## Proof technique 2: greedy stays ahead

Show that after each step, the greedy solution is at least as good as any other solution after the same number of steps. Jump Game: the farthest reachable index after scanning i is the maximum possible; no strategy can reach further.

## Proof technique 3: an invariant that rules out alternatives

Gas Station: if the tank goes negative at station i when starting from s, then no start between s and i works either. Each of them would arrive at i with no more fuel than starting at s did (starting at s, you reached each of them with a non-negative tank). So jump the start to i + 1. One pass. And if total gas ≥ total cost, the final start works, because the deficits before it are covered by the surplus after it.

## Common greedy shapes

- **Reach (Jump Game I and II):** track the farthest reach. Jump Game II counts BFS layers: when i passes the end of the current layer, you must jump, and the new layer ends at the farthest point reached so far.
- **Partition Labels:** a part must extend to the last occurrence of every letter in it. Grow the end to max(last[c]) and cut when i reaches it. Cutting as early as possible maximises the number of parts.
- **Hand of Straights:** the smallest remaining card can only be the start of a group (nothing smaller is left to precede it), so build groups from the smallest card each time. A TreeMap of counts makes this O(n log n).
- **Candy:** split the two-sided constraint into two one-sided passes. Left to right: if rating rises, give one more than the left neighbour. Right to left: if rating falls, ensure one more than the right neighbour (take the max). Each pass gives the minimum for its own constraint; the max of both satisfies both, and nothing smaller would.
- **Task Scheduler:** the most frequent task forces (maxFreq − 1) gaps of size n; other tasks fill gaps, so the answer is max(total tasks, (maxFreq − 1)(n + 1) + number of tasks tied at maxFreq).
- **Assign / match (two sorted lists):** sort both and match smallest to smallest (cookies to children, boats with two people). The exchange argument shows crossing matches never help.

## Greedy vs DP

If a single choice can be justified by an exchange argument, greedy. If the right choice depends on future choices in a way you can't resolve locally, DP. Coin Change with arbitrary coins is the classic: greedy (biggest coin first) fails for coins {1, 3, 4} and amount 6 (greedy 4+1+1, optimal 3+3), so it's DP.

## Common bugs

- No proof, and a counterexample exists.
- Sorting by the wrong key (start time instead of end time for interval scheduling).
- Edge cases: empty input, all equal values, a single element.

## Where it shows up

- **Jump Game I and II, Gas Station, Partition Labels, Hand of Straights, Candy, Task Scheduler.**
- **Non-overlapping Intervals:** earliest end first.
- **IPO, Single-Threaded CPU:** greedy with a heap.
- **Remove K Digits:** greedy with a stack.
