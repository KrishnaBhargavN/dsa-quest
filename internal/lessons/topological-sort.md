---
title: Topological sort and DAGs
group: Graphs
order: 17
topics:
minutes: 12
summary: Order the nodes of a directed graph so every edge points forward; it exists exactly when there is no cycle, and it lets you run DP along edges.
---

## How to recognise it

- Prerequisites, dependencies, build order, "can all courses be finished".
- "Is there a cycle in this directed graph?"
- Longest path, counting paths, or propagating information in a DAG.
- Ordering letters from a sorted alien dictionary.

## Kahn's algorithm (BFS on in-degrees)

```java
int[] indeg = new int[n];
for (int[] e : edges) indeg[e[1]]++;                 // edge e[0] -> e[1]
Deque<Integer> q = new ArrayDeque<>();
for (int i = 0; i < n; i++) if (indeg[i] == 0) q.offer(i);
List<Integer> order = new ArrayList<>();
while (!q.isEmpty()) {
    int u = q.poll();
    order.add(u);
    for (int v : adj.get(u)) if (--indeg[v] == 0) q.offer(v);
}
boolean hasCycle = order.size() < n;
```

**Why it's correct:** a node with in-degree 0 has no unprocessed prerequisites, so placing it next is safe. Removing it can only lower other in-degrees. If a cycle exists, no node on it ever reaches in-degree 0 (each waits on the previous one), so they never get output, which is how `order.size() < n` detects cycles. If there is no cycle, some node always has in-degree 0 (otherwise you could walk backwards forever and would revisit a node: a cycle), so Kahn processes everything.

## DFS order

Run DFS and add a node to a list **after** all its descendants finish (post-order). Reversed, that list is a topological order. **Why:** when u finishes, every node reachable from u has already finished, so u ends up before all of them after reversing.

Cycle detection with DFS needs three colours: 0 = unvisited, 1 = on the current path, 2 = done. Meeting a colour-1 node means a back edge, which means a cycle. Two colours aren't enough in directed graphs: reaching a finished node is not a cycle.

## DP on a DAG

In topological order, every edge goes from an earlier node to a later one, so when you process u, all its predecessors are final. That makes DP along edges straightforward:

- Longest path: dist[v] = max(dist[v], dist[u] + w).
- Count paths: ways[v] += ways[u].
- Largest Color Value: dp[v][c] = max(dp[v][c], dp[u][c]), then add 1 for v's own colour. Kahn's also reports the cycle (return −1).

## Transitive closure (Course Schedule IV)

"Is u a prerequisite of v, directly or indirectly?" for many queries. With n ≤ 100:
- Floyd–Warshall style: reach[i][j] |= reach[i][k] && reach[k][j], with k as the **outer** loop. O(n^3) = 10^6. **Why k outside:** after processing k, reach[i][j] is true iff there is a path using only intermediate nodes from {0..k}. Each round extends the allowed set by one node.
- Or Kahn's order, passing each node's ancestor set (a `BitSet`) to its children.
- Or a BFS/DFS from every node: O(n · (n + e)).

## Alien Dictionary

Compare adjacent words; the first differing character gives an edge a → b. If a word is a prefix of the previous word and shorter, the input is invalid. Then topologically sort the letters.

## Common bugs

- Edge direction reversed: "to take a you need b" is b → a.
- Two-colour cycle detection in a directed graph.
- Forgetting isolated nodes (they have in-degree 0 and must still be output).

## Where it shows up

- **Course Schedule I/II** (you've done these): Kahn's with cycle detection.
- **Course Schedule IV:** transitive closure.
- **Largest Color Value in a Directed Graph:** Kahn's + DP per colour.
- **Minimum Height Trees** (you've done): peeling leaves is Kahn's on an undirected tree.
- **Number of Ways to Arrive at Destination:** Dijkstra's order acts like a topological order of the shortest-path DAG.
