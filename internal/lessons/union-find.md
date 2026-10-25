---
title: Union-find (disjoint set union)
group: Graphs
order: 18
topics:
minutes: 10
summary: Keeps track of which items are connected as you merge groups, in nearly O(1) per operation.
---

## How to recognise it

- Edges arrive one by one and you ask "are these connected now?"
- Counting components while merging.
- Kruskal's MST, detecting a cycle in an undirected graph.
- Grouping by shared attributes (accounts sharing an email, stones sharing a row or column).

## The structure

```java
int[] parent, size;
void init(int n) {
    parent = new int[n]; size = new int[n];
    for (int i = 0; i < n; i++) { parent[i] = i; size[i] = 1; }
}
int find(int x) {
    while (parent[x] != x) {
        parent[x] = parent[parent[x]];   // path halving
        x = parent[x];
    }
    return x;
}
boolean union(int a, int b) {
    a = find(a); b = find(b);
    if (a == b) return false;           // already connected: this edge would close a cycle
    if (size[a] < size[b]) { int t = a; a = b; b = t; }
    parent[b] = a; size[a] += size[b];
    return true;
}
```

Each set is a tree; the root is its representative. Two items are connected iff they have the same root.

**Why it's fast:** union by size keeps trees shallow: a node's depth only increases when its tree is merged into a tree at least as big, so its tree size at least doubles each time, which can happen at most log n times. Path compression (or halving) flattens trees further as you query. Together the amortised cost is O(α(n)), effectively constant.

## Patterns

- **Count components:** start with n, decrement on every successful union.
- **Cycle in an undirected graph:** an edge whose endpoints already share a root closes a cycle (Redundant Connection).
- **Grid problems:** index cells as r * cols + c.
- **Union by attribute:** map each email, row or column to a node id and union items through it.

## Union-find with resets (Find All People With Secret)

Meetings at the same time can pass the secret along a chain instantly, but connections made at time t must not leak into time t + 1. Process meetings grouped by time: union all pairs in the group, then for every person in the group whose root isn't the root of person 0, reset parent[p] = p. Only people in that group were touched, so resetting just them is enough. (An alternative is a BFS per time group.)

## Union-find vs DFS

- Static graph, one question: DFS is simpler.
- Edges added over time with queries in between: union-find.
- Edges **removed** over time: process in reverse as additions.
- You need actual paths: DFS/BFS. Union-find only knows connectivity.

## Common bugs

- Comparing parent[a] == parent[b] instead of find(a) == find(b).
- Recursive find on long chains before compression: stack overflow. Use the iterative version.
- Forgetting to map 2D coordinates or strings to integer ids consistently.

## Where it shows up

- **Accounts Merge, Most Stones Removed** (you've done these).
- **Min Cost to Connect All Points:** Kruskal's needs it (though Prim's is better there).
- **Find All People With Secret:** union-find with resets.
- **Swim in Rising Water:** add cells in height order, stop when start and end connect.
