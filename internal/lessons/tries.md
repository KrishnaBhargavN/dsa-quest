---
title: Tries
group: Trees and recursion
order: 13
topics: Trie
minutes: 10
summary: A tree keyed by characters, so every word sharing a prefix shares a path. Prefix lookups cost O(length of the word), whatever the dictionary size.
---

## How to recognise it

- Many words and questions about **prefixes**: autocomplete, "starts with", replace words by their shortest root.
- Searching for many words at once (in a grid or a stream).
- Wildcard search ('.').
- Maximum XOR of pairs (a binary trie over bits).

## The structure

```java
class Node {
    Node[] next = new Node[26];
    boolean end;          // a word ends here
    String word;          // optional: store the whole word at its end node
}

void insert(Node root, String w) {
    Node cur = root;
    for (char ch : w.toCharArray()) {
        int c = ch - 'a';
        if (cur.next[c] == null) cur.next[c] = new Node();
        cur = cur.next[c];
    }
    cur.end = true;
}
```

`search` walks the same way and checks `end` at the last node; `startsWith` only checks that the walk didn't fall off. Both are O(L) for a word of length L.

**Why it beats a HashSet for prefixes:** a set can tell you whether a whole word exists, but "does any word start with 'ap'?" would need checking every word. In a trie, all words starting with "ap" live under one node, so you answer by walking two edges.

Use `Node[26]` for lowercase letters; use a `HashMap<Character, Node>` for large alphabets.

## Wildcards ('.')

At a '.', try every non-null child recursively. The worst case is exponential in the number of dots, but in practice the trie is sparse.

## Word Search II: backtracking guided by a trie

Searching each word separately repeats the same grid walks. Instead, insert all words into a trie and DFS from each cell, following trie children as you move. You never extend a path that isn't a prefix of some word, which prunes most of the search.

Speed-ups that matter:
- Store the word at its end node; when you reach it, add it and set it to null so it's reported once.
- After exploring a child, if it has no children and no word left, remove it from its parent (prune dead branches).
- Mark the grid cell during the DFS and restore it after.

## Binary tries for XOR

To maximise a XOR b for a fixed a, decide bits from the highest down: at each bit, prefer the child with the **opposite** bit, because a 1 in a higher position beats any combination of lower bits. Insert each number's 31 bits (from bit 30 down) into a trie with two children per node. For each number, walk greedily and build the best XOR. O(31 · n).

**Why greedy is right here:** 2^k > 2^(k−1) + … + 1, so winning a higher bit always beats anything you could gain in lower bits.

## Complexity and memory

Insertion and lookup are O(L). Memory is O(total characters × alphabet size) for array children, which can be large: 10^5 words of length 10 with `Node[26]` is ~26 million references. If memory is tight, use maps or a flattened `int[][]` array.

## Common bugs

- Forgetting the `end` flag (then "app" is "found" after inserting "apple").
- Reporting the same word twice in Word Search II.
- Not restoring grid cells.

## Where it shows up

- **Implement Trie, Design Add and Search Words, Maximum XOR of Two Numbers, Word Search II.**
