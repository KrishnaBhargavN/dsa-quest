---
title: Designing data structures
group: Advanced
order: 34
topics: Design
minutes: 16
summary: "Design X with O(1) operations" means combining structures so each one covers what the others can't. Start from the operations, then pick a structure per operation.
---

## How to recognise it

- "Implement a class with methods get / put / add / remove / getRandom…" plus complexity targets.
- LRU, LFU, time-based stores, underground systems, Twitter, browser history, hit counters.

## The method

1. List every operation and its required complexity.
2. For each, ask which structure gives it: O(1) lookup by key → HashMap; order by recency → doubly linked list; min/max → heap; ordered keys and floor/ceiling → TreeMap; random access → ArrayList.
3. Combine them and decide what each one stores (often the map points **into** the other structure).
4. Walk through each operation and update **every** structure. Most design bugs are one structure falling out of sync.

## Java collections cheat sheet

| Structure | Good at | Cost |
| --- | --- | --- |
| HashMap / HashSet | lookup, insert, delete by key | O(1) average |
| TreeMap / TreeSet | ordered keys, floorKey, ceilingKey, first, last | O(log n) |
| ArrayList | index access, append | O(1); middle insert/delete O(n) |
| ArrayDeque | stack and queue at both ends | O(1) |
| PriorityQueue | min (or max) | O(log n) insert/poll, O(n) arbitrary remove |
| LinkedHashMap | insertion or access order | O(1) |

## LRU Cache

Operations: get(key) and put(key, value), both O(1); evict the least recently used when full.

- HashMap key → node, for O(1) lookup.
- Doubly linked list of nodes ordered by recency, for O(1) "move to front" and "remove from back". Dummy head and tail remove null checks.

```java
class LRUCache {
    class Node { int key, val; Node prev, next; Node(int k, int v) { key = k; val = v; } }
    private final Map<Integer, Node> map = new HashMap<>();
    private final Node head = new Node(0, 0), tail = new Node(0, 0);
    private final int cap;
    LRUCache(int capacity) { cap = capacity; head.next = tail; tail.prev = head; }

    private void unlink(Node n) { n.prev.next = n.next; n.next.prev = n.prev; }
    private void addFront(Node n) { n.next = head.next; n.prev = head; head.next.prev = n; head.next = n; }

    public int get(int key) {
        Node n = map.get(key);
        if (n == null) return -1;
        unlink(n); addFront(n);
        return n.val;
    }
    public void put(int key, int value) {
        Node n = map.get(key);
        if (n != null) { n.val = value; unlink(n); addFront(n); return; }
        if (map.size() == cap) { Node lru = tail.prev; unlink(lru); map.remove(lru.key); }
        n = new Node(key, value); map.put(key, n); addFront(n);
    }
}
```

The node stores its key so eviction can remove it from the map. (Java's `LinkedHashMap` with access order and `removeEldestEntry` does all of this, but interviewers want to see the structure.)

## LFU Cache

Evict the least frequently used; break ties by least recently used. All O(1).

- key → (value, freq).
- freq → an ordered set of keys at that frequency, in recency order (`LinkedHashSet`, or a doubly linked list per frequency).
- minFreq: the smallest frequency currently present.

On access: move the key from freq f to f + 1; if f's set becomes empty and f == minFreq, minFreq++. On insert: freq 1 and minFreq = 1. On evict: remove the oldest key in minFreq's set.

**Why minFreq only ever needs +1 or a reset to 1:** an access moves one key up by exactly one level, so the minimum can rise by at most one; an insert creates frequency 1, the lowest possible.

## Insert Delete GetRandom O(1)

An ArrayList gives O(1) random access; a HashMap value → index gives O(1) lookup. To remove from the middle of the list in O(1), move the **last** element into the hole, update its index in the map, and remove the last slot. Order doesn't matter for a set, so this is allowed.

## Time-based and ordered data

- **Time Based Key-Value Store:** key → list of (timestamp, value), appended in increasing time; binary search for the last timestamp ≤ t. Or key → TreeMap with floorEntry.
- **Underground System:** check-in map id → (station, time); route map "from,to" → (total, count). Averages are computed on demand.
- **Design Twitter:** user → list of tweets with a global timestamp; feed = k-way merge of followees' lists with a heap, stopping after 10. Store follows as user → HashSet.
- **Hit Counter / rate limiter:** a queue of timestamps, dropping those older than the window.

## Thinking aloud in the interview

State the operations and target complexities first, propose the structures, then code. Mention trade-offs (memory, concurrency, what happens at capacity 0). It shows design judgement, not just code.

## Common bugs

- Updating one structure but not the other (the classic LRU bug: forgetting map.remove on eviction).
- Not storing the key in the list node.
- Removing from a list while iterating it.
- LFU: forgetting to update minFreq when a set empties.

## Where it shows up

- **LRU Cache, LFU Cache, Insert Delete GetRandom O(1), Design Underground System, Time Based Key-Value Store, Design Twitter, Find Median from Data Stream, Implement Trie, Design Add and Search Words.**
