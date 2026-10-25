---
title: String algorithms: KMP, Z and rolling hash
group: Advanced
order: 33
topics:
minutes: 16
summary: Find a pattern in text in O(n + m) instead of O(n · m), detect repeats and borders, and compare substrings in O(1) with hashing.
---

## How to recognise it

- "Find all occurrences of a pattern", "first occurrence" (strStr) with large inputs.
- "Shortest palindrome by adding characters in front", "repeated substring pattern", "longest prefix that is also a suffix".
- "Longest duplicate substring", comparing many substrings for equality.

For small inputs, `text.indexOf(pattern)` is fine and interviewers often accept it; know KMP for when they ask "can you do better?"

## Java string basics

- Strings are immutable; build with `StringBuilder`.
- `s.charAt(i)` is fine; `s.toCharArray()` is faster in tight loops.
- `substring` copies (O(length)); avoid it inside loops.
- `s1.equals(s2)`, never `==`.

## KMP: the prefix function

pi[i] = length of the longest proper prefix of s[0..i] that is also a suffix of it.

```java
int[] prefixFunction(String s) {
    int n = s.length();
    int[] pi = new int[n];
    for (int i = 1; i < n; i++) {
        int k = pi[i - 1];
        while (k > 0 && s.charAt(i) != s.charAt(k)) k = pi[k - 1];
        if (s.charAt(i) == s.charAt(k)) k++;
        pi[i] = k;
    }
    return pi;
}
```

**Why the fallback `k = pi[k − 1]` is right:** if the border of length k can't be extended by s[i], the next candidate must be a shorter border of the same prefix, and the longest shorter border of s[0..k−1] is exactly pi[k − 1]. You never skip a candidate.

**Why it's O(n):** k increases by at most 1 per step and every fallback decreases it, so the total number of fallbacks is at most n.

**Pattern search:** compute pi on `pattern + "#" + text` (with a separator that appears in neither). Every position where pi equals the pattern length is a match ending there. O(n + m).

Uses beyond search:
- **Repeated substring pattern:** with p = pi[n−1], the string repeats a block iff p > 0 and n % (n − p) == 0.
- **Shortest palindrome:** the longest palindromic prefix of s is pi's last value for s + "#" + reverse(s).

## Z-function

z[i] = length of the longest substring starting at i that matches a prefix of s. Computed in O(n) by reusing the rightmost match window [l, r]. Pattern search: compute z on pattern + "#" + text and look for z[i] == m. Equivalent in power to KMP; some find it easier to reason about.

## Rolling hash (Rabin-Karp)

Treat a string as a number in base B mod M: h(s) = s[0]·B^(m−1) + … + s[m−1]. With prefix hashes H and powers P precomputed, the hash of any substring s[l..r) is H[r] − H[l]·P[r − l] (mod M), in O(1).

```java
long M = 1_000_000_007L, B = 131;
long[] H = new long[n + 1], P = new long[n + 1];
P[0] = 1;
for (int i = 0; i < n; i++) {
    H[i + 1] = (H[i] * B + s.charAt(i)) % M;
    P[i + 1] = P[i] * B % M;
}
long hash(int l, int r) { return ((H[r] - H[l] * P[r - l]) % M + M) % M; }
```

Equal strings always have equal hashes; unequal strings collide with probability about 1/M per comparison. For safety use two moduli (or verify matches directly).

**Longest duplicate substring:** binary search the length L (if a duplicate of length L exists, so does one of L − 1), and check each L with a HashSet of substring hashes: O(n log n).

## Palindromes in O(n)

Manacher's algorithm finds the longest palindrome around every centre in O(n). It's rarely required; expand-around-centre (O(n^2)) is the expected interview answer.

## Common bugs

- No separator in `pattern + "#" + text`, so matches can span the boundary.
- Hash subtraction going negative; add M before the final mod.
- Using `substring` in a loop and turning O(n) into O(n^2).

## Where it shows up

- Not on your list directly, but common in interviews: Find the Index of the First Occurrence, Repeated Substring Pattern, Shortest Palindrome, Longest Happy Prefix (pure prefix function), Longest Duplicate Substring.
- **Word Search II and Design Add and Search Words** use tries instead (Tries lesson).
