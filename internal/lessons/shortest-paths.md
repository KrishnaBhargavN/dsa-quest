---
title: Shortest paths: Dijkstra, Bellman-Ford, Floyd-Warshall
group: Graphs
order: 19
topics:
minutes: 15
summary: Pick the algorithm by edge weights and graph size: BFS for equal weights, Dijkstra for non-negative, Bellman-Ford for negative or limited hops, Floyd-Warshall for all pairs.
---

## Choosing the algorithm

| Situation | Algorithm | Time |
| --- | --- | --- |
| All edges cost the same | BFS | O(V + E) |
| Costs 0 or 1 | 0-1 BFS | O(V + E) |
| Non-negative weights | Dijkstra | O(E log V) |
| Negative weights, or "at most k edges" | Bellman-Ford | O(V · E) or O(k · E) |
| All pairs, V ≤ ~400 | Floyd-Warshall | O(V^3) |

## Dijkstra

```java
long[] dist = new long[n];
Arrays.fill(dist, Long.MAX_VALUE);
dist[src] = 0;
PriorityQueue<long[]> pq = new PriorityQueue<>((a, b) -> Long.compare(a[0], b[0]));
pq.offer(new long[]{0, src});
while (!pq.isEmpty()) {
    long[] cur = pq.poll();
    long d = cur[0]; int u = (int) cur[1];
    if (d > dist[u]) continue;                       // stale entry
    for (int[] e : adj.get(u)) {
        int v = e[0]; long nd = d + e[1];
        if (nd < dist[v]) { dist[v] = nd; pq.offer(new long[]{nd, v}); }
    }
}
```

**Why it's correct (non-negative weights):** when u is popped with the smallest tentative distance d, any other route to u would have to leave the set of finished nodes through some node x with tentative distance ≥ d, and every edge after that adds ≥ 0. So nothing can beat d; u is final. A negative edge breaks exactly this step, which is why Dijkstra fails with negative weights.

The `if (d > dist[u]) continue` line replaces a decrease-key operation: old entries stay in the heap but are skipped.

## Variations

- **Count shortest paths:** when nd < dist[v], set ways[v] = ways[u]; when nd == dist[v], add ways[u]. Because nodes are finalised in distance order, ways[u] is complete when u is popped. Use `long` distances and mod the counts.
- **Minimise the maximum edge on the path** (Swim in Rising Water, Path With Minimum Effort): replace d + w with max(d, w). Dijkstra still works because max is also non-decreasing along a path.
- **Multiple sources:** push them all at distance 0.
- **States:** the node can be a tuple, like (city, stops used) or (node, fuel).

## Bellman-Ford

Relax every edge V − 1 times. **Why V − 1:** a shortest path without cycles uses at most V − 1 edges, and after round i every shortest path with ≤ i edges is correct (induction on i). One more round that still improves something proves a negative cycle is reachable.

**At most k stops (Cheapest Flights):** run k + 1 rounds, and in each round relax from a **copy** of the previous round's distances. Without the copy, one round could chain several edges and use more hops than allowed.

```java
for (int round = 0; round <= k; round++) {
    long[] next = dist.clone();
    for (int[] f : flights)
        if (dist[f[0]] != INF && dist[f[0]] + f[2] < next[f[1]]) next[f[1]] = dist[f[0]] + f[2];
    dist = next;
}
```

## Floyd-Warshall

```java
for (int k = 0; k < n; k++)
    for (int i = 0; i < n; i++)
        for (int j = 0; j < n; j++)
            if (d[i][k] + d[k][j] < d[i][j]) d[i][j] = d[i][k] + d[k][j];
```

**Why k is the outer loop:** after iteration k, d[i][j] is the shortest path using only intermediate nodes 0..k. The new option at step k is "go through k", combining two paths that use only 0..k−1. Put k inside and the invariant breaks. Use a large INF that won't overflow when added (like 1e15 for long, or check for INF first).

## Common bugs

- `int` distances overflowing; use `long`.
- Marking nodes visited when pushed (that's BFS logic) instead of skipping stale entries when popped.
- Using Dijkstra with negative edges.
- Bellman-Ford with k hops without the copy array.

## Where it shows up

- **Network Delay Time, Cheapest Flights Within K Stops, Path With Minimum Effort** (you've done these).
- **Number of Ways to Arrive at Destination:** Dijkstra + counting.
- **Swim in Rising Water:** Dijkstra with max instead of sum.
- **Evaluate Division** (done): weighted graph, multiply along paths.
- **Minimum Cost to Make a Valid Path:** 0-1 BFS (see Graph traversal).
