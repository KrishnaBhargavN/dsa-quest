---
title: Arrays and hashing
group: Foundations
order: 2
topics: Arrays and hashing
minutes: 12
summary: Trade memory for time: a hash map turns "have I seen this?" from O(n) into O(1).
---

## How to recognise it

- You keep asking "have I seen x before?" or "how many times have I seen x?"
- Grouping things that are "the same" under some transformation (anagrams, same row, same difference).
- The brute force is two nested loops comparing pairs.

## The core idea

Most pair-search problems are "for each element, find a partner". The brute force scans for the partner: O(n^2). A hash map remembers everything already scanned, so finding the partner is O(1): O(n) total.

The shape is always:

```java
Map<Integer, Integer> seen = new HashMap<>();
for (int i = 0; i < nums.length; i++) {
    int need = target - nums[i];
    if (seen.containsKey(need)) return new int[]{seen.get(need), i};
    seen.put(nums[i], i);
}
```

**Why it works:** when you are at index i, the map holds exactly the elements before i. Every pair (j, i) with j < i is checked at the moment you reach i, so every pair is checked exactly once and none is missed.

## Canonical keys (grouping)

To group "equivalent" items, map each one to a **canonical form** and use it as the key. Two items are equivalent if and only if their canonical forms are equal, so grouping by key groups exactly the equivalent items.

- Anagrams: sorted characters (`"eat"` → `"aet"`), or a 26-count signature (`"1#0#0#…"`). Sorting is O(k log k) per word; counting is O(k).
- Points on the same line through the origin: the reduced fraction dy/dx (divide by gcd, fix the sign).

```java
Map<String, List<String>> groups = new HashMap<>();
for (String w : words) {
    char[] c = w.toCharArray();
    Arrays.sort(c);
    groups.computeIfAbsent(new String(c), k -> new ArrayList<>()).add(w);
}
```

## Counting with a map

`map.merge(x, 1, Integer::sum)` increments a count. For lowercase letters, an `int[26]` is faster and simpler than a map.

## Sets for O(1) membership

Longest Consecutive Sequence is the classic: put everything in a `HashSet`. A number x **starts** a run only if x − 1 is absent. From each start, walk x+1, x+2… while present.

**Why it's O(n) even though there is a nested loop:** the inner walk only runs from run starts, and each number belongs to exactly one run, so the inner loop visits every number at most once in total.

## Prefix and suffix products

Product of Array Except Self: answer[i] = (product of everything left of i) × (product of everything right of i). Fill left products in one pass, then multiply in right products with a running variable going backwards. No division, O(1) extra space besides the output. The same "left part × right part" decomposition shows up in Trapping Rain Water and Candy.

## Prefix sums meet hashing

"Count subarrays with sum k" with negatives allowed can't use a sliding window. Use prefix sums: sum(i..j) = P[j] − P[i−1]. You want earlier prefixes equal to P[j] − k, so count them in a map as you go. Seed the map with {0: 1} so subarrays starting at index 0 count. See the Prefix sums lesson for the full argument.

## Matrix tricks worth knowing

- Rotate 90° clockwise in place: transpose, then reverse each row.
- Spiral order: keep four boundaries (top, bottom, left, right) and shrink them after each side.
- Set matrix zeroes in O(1) space: use the first row and column as marker storage, with one extra flag for the first column.

## Common bugs

- Using `int[]` or `char[]` as a map key: arrays use identity equality. Convert to `String` or `List`.
- `Integer` comparison with `==` works only for −128..127 (cached). Use `.equals` or unbox.
- Forgetting the {0: 1} seed in prefix-sum counting.
- Mutating a list that is also a key in a map.

## Where it shows up

- **Group Anagrams:** canonical key.
- **Longest Consecutive Sequence:** set plus "start of run" check.
- **Product of Array Except Self:** prefix × suffix.
- **Subarray Sum Equals K:** prefix sums + count map. Also Path Sum III, the same trick on a tree.
