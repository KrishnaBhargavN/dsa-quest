---
title: Bridges, SCCs, bipartite graphs and Euler paths
group: Graphs
order: 21
topics:
minutes: 16
summary: The less common graph tools: Tarjan's low-link for bridges and strongly connected components, two-colouring for bipartite checks, and Hierholzer for using every edge once.
---

## How to recognise it

- "Critical connections", "which edge removal disconnects the network": bridges.
- "Which node removal disconnects it": articulation points.
- "Groups where everyone can reach everyone" in a directed graph: SCCs.
- "Split into two groups with no edge inside a group", "possible bipartition": bipartite check.
- "Use every ticket / edge exactly once": Euler path (Hierholzer).

## Discovery time and low-link

Run DFS and stamp each node with `tin[u]`, the time it was first visited. Define

low[u] = the smallest tin reachable from u's DFS subtree using tree edges down and **at most one back edge** up.

```java
int timer = 0;
void dfs(int u, int parentEdge) {
    tin[u] = low[u] = timer++;
    for (int[] e : adj.get(u)) {                      // e = {to, edgeId}
        int v = e[0];
        if (e[1] == parentEdge) continue;             // don't reuse the edge we came in on
        if (tin[v] != -1) {
            low[u] = Math.min(low[u], tin[v]);        // back edge
        } else {
            dfs(v, e[1]);
            low[u] = Math.min(low[u], low[v]);
            if (low[v] > tin[u]) bridges.add(new int[]{u, v});
        }
    }
}
```

**Why `low[v] > tin[u]` means (u, v) is a bridge:** low[v] > tin[u] says nothing in v's subtree can climb back to u or above without using the edge (u, v). Remove it and the subtree is cut off. If some back edge reached u or an ancestor, that would be a second route, and the edge wouldn't be a bridge.

Skipping by **edge id** (not parent node) handles parallel edges correctly.

**Articulation points:** a non-root u is one if some child v has low[v] ≥ tin[u] (≥, because v's subtree can reach u but not above it, so removing u cuts it off). The root is one if it has two or more DFS children.

For 10^5 nodes, recursion depth can overflow Java's default stack; run the DFS in a `new Thread(null, runnable, "dfs", 1 << 26)` or convert it to iterative.

## Strongly connected components

In a directed graph, an SCC is a maximal set where every node reaches every other.

**Kosaraju (easiest to get right):**
1. DFS on the graph; push nodes to a stack in post-order (finish time).
2. Reverse every edge.
3. Pop nodes from the stack; each DFS in the reversed graph from an unvisited node collects one SCC.

**Why it works:** the node that finishes last is in a "source" SCC of the condensed DAG. In the reversed graph that SCC becomes a sink, so a DFS from it can't leak into other components. Removing it and repeating peels the components in order.

**Tarjan** does it in one DFS with low-links and an explicit stack: when low[u] == tin[u], pop the stack down to u; that's one SCC. Same idea as bridges, but low only counts nodes still on the stack.

Contracting SCCs gives a DAG, where topological sort and DAG DP apply.

## Bipartite check

Colour a start node 0, give every neighbour the opposite colour (BFS or DFS). If an edge ever joins two same-coloured nodes, it isn't bipartite. Run from every uncoloured node (the graph may be disconnected).

**Why:** a graph is bipartite iff it has no odd cycle. Alternate colouring along any path assigns colours by distance parity, so a conflict happens exactly when two routes to a node have different parities, which means an odd cycle.

## Euler paths (Hierholzer)

A path that uses every edge exactly once exists in a directed graph when at most one node has out − in = 1 (the start), at most one has in − out = 1 (the end), the rest are balanced, and all edges are connected.

```java
void visit(String u) {
    PriorityQueue<String> out = graph.get(u);         // smallest destination first
    while (out != null && !out.isEmpty()) visit(out.poll());
    route.addFirst(u);                                // post-order, prepended
}
```

**Why post-order:** you may walk into a dead end before using every edge. The dead end must be the **last** stop of the final route, so you add nodes only when they have no unused edges left; the remaining loops get spliced in before them. Using a min-heap of destinations makes the route lexicographically smallest (Reconstruct Itinerary).

## Common bugs

- Bridges: skipping the parent **node** instead of the parent **edge** (breaks with parallel edges).
- Using `low[v] >= tin[u]` for bridges (that's the articulation point condition).
- Recursion depth on large graphs.
- Hierholzer: adding the node before exploring its edges (pre-order) gives a wrong route.

## Where it shows up

- **Critical Connections in a Network** (you solved it first try): bridges.
- **Is Graph Bipartite, Possible Bipartition** (done): two-colouring.
- **Reconstruct Itinerary:** Hierholzer with a min-heap.
- SCCs rarely appear directly on LeetCode but come up in system-style questions ("find mutually dependent services").
