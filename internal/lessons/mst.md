---
title: Minimum spanning trees
group: Graphs
order: 20
topics:
minutes: 10
summary: Connect every node at minimum total cost. Kruskal for sparse edge lists, Prim for dense or complete graphs; both rest on the cut property.
---

## How to recognise it

- "Connect all points / cities / houses with minimum total cost."
- Every node must be reachable, there is no fixed source, and you pay per edge used.
- Contrast with shortest paths, which minimise the cost from one source, not the total.

## The cut property (why both algorithms work)

Split the nodes into any two groups. Among the edges crossing between the groups, the cheapest one belongs to some minimum spanning tree.

**Proof sketch:** take an MST that doesn't use that cheapest crossing edge e. Add e: it forms a cycle, and that cycle must cross the cut a second time through some edge f. f costs at least as much as e, so swapping f for e gives a spanning tree that costs no more. So some MST contains e.

Both algorithms repeatedly add a cheapest edge across some cut, so both are correct.

## Kruskal

Sort edges by weight, add each edge unless it joins two nodes that are already connected (union-find). Stop after n − 1 edges.

**Why it's the cut property:** when Kruskal adds edge (u, v), take the cut "u's current component vs everything else". Every cheaper edge was already considered, so (u, v) is the cheapest edge crossing this cut.

O(E log E) for the sort. Best when you have an explicit, sparse edge list.

## Prim

Grow one tree from any start node; repeatedly add the cheapest edge from the tree to a node outside it.

For a **complete graph** (every pair of points connected, like Min Cost to Connect All Points), skip the heap and edge list entirely:

```java
int n = pts.length;
int[] best = new int[n];              // cheapest known edge from the tree to each node
Arrays.fill(best, Integer.MAX_VALUE);
boolean[] in = new boolean[n];
best[0] = 0;
int total = 0;
for (int round = 0; round < n; round++) {
    int u = -1;
    for (int i = 0; i < n; i++) if (!in[i] && (u == -1 || best[i] < best[u])) u = i;
    in[u] = true;
    total += best[u];
    for (int v = 0; v < n; v++) {
        if (in[v]) continue;
        int w = Math.abs(pts[u][0] - pts[v][0]) + Math.abs(pts[u][1] - pts[v][1]);
        if (w < best[v]) best[v] = w;
    }
}
```

O(n^2) time and O(n) memory, versus Kruskal's O(n^2 log n) time and O(n^2) memory just to list the edges. With n = 1000, that's 10^6 simple steps.

With a heap (sparse graphs), Prim is O(E log V), like Dijkstra but keyed on the edge weight instead of the distance from the source.

## Common bugs

- Confusing Prim with Dijkstra: Prim uses the **edge** weight w, Dijkstra uses dist[u] + w.
- Building all n^2 edges for a complete graph and running out of memory or time.
- Kruskal without checking that n − 1 edges were added (the graph may be disconnected).

## Where it shows up

- **Min Cost to Connect All Points:** O(n^2) Prim.
- **Swim in Rising Water** is close: adding cells in height order with union-find is Kruskal's idea (the answer is the largest edge on the minimum spanning path).
