---
title: String DP
group: Dynamic programming
order: 25
topics: DP - strings
minutes: 16
summary: Two strings? The state is a pair of prefixes, dp[i][j], and the transition looks at the last characters. One string? Usually an interval dp[i][j] over s[i..j].
---

## How to recognise it

- Two strings and "common", "edit", "transform", "match", "interleave", "how many times does t appear in s".
- One string and "palindromic subsequence / substring", "partition into palindromes".
- Patterns with wildcards ('.', '*', '?').

## The prefix-pair state

dp[i][j] = the answer for the first i characters of s and the first j characters of t. Sizes (m + 1) × (n + 1), so row 0 and column 0 are the empty prefixes; base cases live there.

**Why prefixes work:** any alignment of s[0..i) and t[0..j) ends in one of a few ways: the last characters match each other, or one of them is unmatched. Each case leaves a smaller pair of prefixes, which is a subproblem.

## LCS

```java
int[][] dp = new int[m + 1][n + 1];
for (int i = 1; i <= m; i++)
    for (int j = 1; j <= n; j++)
        dp[i][j] = s.charAt(i - 1) == t.charAt(j - 1)
            ? dp[i - 1][j - 1] + 1
            : Math.max(dp[i - 1][j], dp[i][j - 1]);
```

**Why taking a matching pair is safe:** if the last characters are equal, some LCS uses them as its last pair. If an LCS didn't, you could swap its last pair for this one without losing length.

## Edit distance

dp[i][j] = fewest operations to turn s[0..i) into t[0..j).
- Last characters equal: dp[i−1][j−1].
- Otherwise 1 + min of: replace dp[i−1][j−1], delete from s dp[i−1][j], insert into s dp[i][j−1].
- Base: dp[i][0] = i (delete everything), dp[0][j] = j (insert everything).

## Counting: Distinct Subsequences

dp[i][j] = ways for s[0..i) to produce t[0..j). Character s[i−1] can be skipped (dp[i−1][j]) or, if it equals t[j−1], used (dp[i−1][j−1]). dp[i][0] = 1: one way to make the empty string. In 1D, loop j **downward** so dp[j−1] is still the previous row's value (same reason as 0/1 knapsack).

## Interleaving String

dp[i][j] = can s1[0..i) and s2[0..j) interleave into s3[0..i+j). The next character of s3 is at index i + j − 1 and must come from s1's or s2's last character. Check lengths first.

## Regular Expression Matching

dp[i][j] = does s[0..i) match p[0..j)?
- p[j−1] is a letter or '.': it must match s[i−1], and dp[i−1][j−1].
- p[j−1] is '*' (applies to p[j−2]): either use it zero times, dp[i][j−2], or, if p[j−2] matches s[i−1], use it once more and stay on the same pattern position, dp[i−1][j].
- Base: dp[0][0] = true; dp[0][j] = true when p[0..j) is like a*b*c* (each '*' can erase its letter): dp[0][j] = p[j−1] == '*' && dp[0][j−2].

**Why "stay on j" for the star:** x* can absorb any number of x's. Consuming one s character but keeping the same pattern state lets it absorb another one next time.

## One string: palindromes as intervals

- **Longest Palindromic Subsequence:** dp[i][j] over s[i..j]: if s[i] == s[j], 2 + dp[i+1][j−1]; else max(dp[i+1][j], dp[i][j−1]). Fill i from the end (i depends on i + 1). It also equals LCS(s, reverse(s)).
- **Longest Palindromic Substring:** expand around each of the 2n − 1 centres: O(n^2) time, O(1) space, simpler than DP.
- **isPal table:** isPal[i][j] = s[i] == s[j] && (j − i < 2 || isPal[i+1][j−1]). Reusable for partitioning problems.

## Space

When dp[i][·] only reads row i − 1, keep two rows. Keep one row only when you can order the loop so the values you read are still the old ones (or save the diagonal in a variable).

## Common bugs

- Off-by-one between dp indices (prefix length) and string indices (i − 1).
- Missing base row/column initialisation.
- In regex: forgetting the dp[0][j] row for patterns that match the empty string.
- Filling interval DP in the wrong order (i must go from n − 1 down).

## Where it shows up

- **Longest Common Subsequence, Edit Distance, Longest Palindromic Substring, Longest Palindromic Subsequence, Interleaving String, Distinct Subsequences, Regular Expression Matching.**
- **Palindrome Partitioning:** backtracking over an isPal table.
