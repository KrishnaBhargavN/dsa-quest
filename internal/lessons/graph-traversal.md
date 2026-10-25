---
title: Graph traversal: BFS, DFS and grids
group: Graphs
order: 16
topics: Graphs
minutes: 16
summary: BFS finds shortest paths when every step costs the same; DFS explores and labels components. Most graph problems are one of these with the right notion of "state".
---

## Which graph lesson do I need?

This is the starting lesson for every graph problem. Pick the follow-up by what the problem asks:
- Order of tasks, prerequisites, cycles in a directed graph → **Topological sort and DAGs**.
- Merging groups as edges arrive → **Union-find**.
- Weighted shortest path → **Shortest paths**.
- Connect everything at minimum total cost → **Minimum spanning trees**.
- Critical edges, strongly connected groups, two-colouring, using every edge once → **Bridges, SCCs, bipartite graphs and Euler paths**.

## How to recognise it

- Grids with cells you can move between, islands, regions, flood fill.
- "Minimum number of steps / moves / transformations" where every move costs the same: BFS.
- "Connected", "reachable", "count the groups": DFS, BFS or union-find.
- Things spreading simultaneously (rotting, fire, infection): multi-source BFS.

## Representing graphs in Java

```java
List<List<Integer>> adj = new ArrayList<>();
for (int i = 0; i < n; i++) adj.add(new ArrayList<>());
for (int[] e : edges) { adj.get(e[0]).add(e[1]); adj.get(e[1]).add(e[0]); }
```

Weighted edges: `List<List<int[]>>` with `{to, weight}`. Grids need no adjacency list; use direction arrays:

```java
int[][] DIRS = {{1,0},{-1,0},{0,1},{0,-1}};
for (int[] d : DIRS) {
    int nr = r + d[0], nc = c + d[1];
    if (nr < 0 || nc < 0 || nr >= m || nc >= n) continue;
    // ...
}
```

## BFS and why it gives shortest paths

```java
int[] dist = new int[n];
Arrays.fill(dist, -1);
Deque<Integer> q = new ArrayDeque<>();
dist[src] = 0; q.offer(src);
while (!q.isEmpty()) {
    int u = q.poll();
    for (int v : adj.get(u)) {
        if (dist[v] != -1) continue;
        dist[v] = dist[u] + 1;
        q.offer(v);
    }
}
```

**Why it's correct:** the queue always holds nodes at distance d followed by nodes at distance d + 1, never anything else. So nodes leave the queue in order of distance, and the first time you reach v is via a shortest path. Mark visited **when you enqueue**, not when you dequeue, or the same node can be queued many times.

Level by level: read `q.size()` at the start of each round and process exactly that many nodes. That gives you the level number for free (minutes in Rotting Oranges, steps in Word Ladder).

## Multi-source BFS

Put **all** sources in the queue at distance 0 before starting. It's the same as BFS from a virtual super-source connected to every source, so dist[v] becomes the distance to the **nearest** source. Rotting Oranges, "distance to nearest 0", walls and gates.

## BFS over states, not just positions

When position alone doesn't describe the situation, make the state a tuple and BFS over tuples:
- Grid with k obstacle removals: (r, c, removalsLeft).
- Visit all nodes: (node, visitedMask).
- Keys and doors: (r, c, keysMask).

visited must be indexed by the whole state. The state count is the product of the parts, so check it fits.

## 0-1 BFS

If edges cost 0 or 1, use a deque: push 0-cost neighbours to the **front** and 1-cost neighbours to the **back**. The deque stays sorted by distance (at most two distinct values, d and d + 1), so it behaves like Dijkstra in O(V + E). Minimum Cost to Make a Valid Path in a Grid is exactly this.

## DFS

```java
void dfs(int u, boolean[] seen) {
    seen[u] = true;
    for (int v : adj.get(u)) if (!seen[v]) dfs(v, seen);
}
```

Recursive DFS on a 10^5-node path overflows Java's stack. For big inputs, use an explicit stack, or BFS when order doesn't matter.

Counting components: loop over all nodes and start a DFS from each unseen one; each start is a new component.

## Labelling components (Making A Large Island)

Give each island an id with one DFS per island and store size[id]. Then for each 0, look at its four neighbours, collect the **distinct** ids (a set of up to 4), and sum their sizes plus 1. Distinct matters: two sides of the same island must not be counted twice. Each cell is labelled once, so it's O(n^2) overall.

## Reverse the question

"Which cells can reach the ocean?" is expensive cell by cell. "Which cells can the ocean reach going uphill?" is one multi-source search from the border. Pacific Atlantic Water Flow runs it from both oceans and intersects. Surrounded Regions does the same from the border O's.

## Implicit graphs

The graph often isn't given. In Word Ladder, words are nodes and one-letter changes are edges. Generate neighbours on the fly: for each position try 26 letters and check a HashSet. Remove a word from the set when you visit it so it's never re-queued.

## Common bugs

- Marking visited on dequeue instead of enqueue.
- Forgetting the grid bounds check before reading the cell.
- Reusing a visited array across test cases or calls (the shared-state bug).
- Using DFS for a shortest path in an unweighted graph. DFS doesn't give shortest paths.

## Where it shows up

- **Making A Large Island:** component labelling.
- **Rotting Oranges:** multi-source BFS by levels.
- **Pacific Atlantic Water Flow:** reverse search from both borders.
- **Word Ladder:** BFS over an implicit graph.
- **Shortest Path with Obstacles Elimination:** BFS over (r, c, k).
- **Shortest Path Visiting All Nodes:** BFS over (node, mask).
- **Minimum Cost to Make a Valid Path:** 0-1 BFS.
- **All Nodes Distance K:** trees are graphs too; add parent links and BFS.
