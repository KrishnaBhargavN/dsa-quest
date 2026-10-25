---
title: DP on trees
group: Dynamic programming
order: 27
topics: DP - trees
minutes: 11
summary: Do a post-order DFS where each node returns a small tuple of answers, one per state of the node (taken or not, covered or not), and combine the children's tuples.
---

## How to recognise it

- A tree, and a choice at each node that constrains its neighbours: "can't pick a node and its child", "every node must be covered", "choose a subset with some property".
- Asking for an optimum or a count over all such choices.

## The core idea

A subtree's best answer depends on only a little information about its root: was the root picked, is it covered, does it hold a camera. So let the DFS return **one value per root state**, and let the parent combine its children's tuples. That's post-order: children first, then the parent.

**Why it's correct:** the constraint between a node and the rest of the tree only goes through the node itself. Once you know the root's state, the best choices inside each child's subtree are independent of everything else. That's optimal substructure on a tree.

## House Robber III

```java
// returns {best if root is robbed, best if root is not robbed}
long[] dfs(TreeNode node) {
    if (node == null) return new long[]{0, 0};
    long[] l = dfs(node.left), r = dfs(node.right);
    long rob = node.val + l[1] + r[1];                          // children must be skipped
    long skip = Math.max(l[0], l[1]) + Math.max(r[0], r[1]);     // children are free
    return new long[]{rob, skip};
}
```

The answer is max of the root's two values. Plain memoised recursion over "grandchildren" also works but is slower and messier; the tuple is the clean version.

## Binary Tree Cameras: three states

Each node, after its subtree is handled, is in one of three states:
- 0: not covered (its parent must cover it),
- 1: has a camera,
- 2: covered by a child, no camera.

Rules, bottom-up:
- If any child is 0, this node **must** get a camera (only it can still cover that child): return 1.
- Else if any child has a camera, this node is covered: return 2.
- Else (children covered but no cameras nearby): return 0 and let the parent handle it.
- A null child counts as 2 (covered, no camera), so leaves return 0, and their parents get cameras.

At the end, if the root returns 0, add one more camera.

**Why this greedy-looking DP is optimal:** a camera on a leaf covers at most 2 nodes; moving it up to the leaf's parent covers that leaf, the parent, the parent's other child and the grandparent: at least as much. So never put cameras on leaves, and only add a camera where it's forced. This is an exchange argument applied bottom-up.

## Rerooting (all roots at once)

Some problems need the answer for **every** node as the root ("sum of distances to all nodes"). Compute the answer for one root with a DFS, then a second DFS moves the root to each child, adjusting the answer in O(1): when moving from u to child v, the size(v) nodes in v's subtree get 1 closer and the other n − size(v) get 1 farther, so ans[v] = ans[u] − size(v) + (n − size(v)). O(n) total.

## Diameter and longest paths

Diameter: each node returns its height; the best path through the node is left height + right height; keep a global max. Binary Tree Maximum Path Sum (Trees lesson) is the same pattern with sums and "drop negative branches".

## Common bugs

- Returning a single number when the parent needs to know the child's state.
- A global or static counter not reset between test cases.
- Forgetting the final check at the root (Cameras).
- Recursion depth on very deep trees.

## Where it shows up

- **House Robber III, Binary Tree Cameras.**
- **Binary Tree Maximum Path Sum:** return one value up, keep a global for the bent path.
- **Minimum Height Trees** (you've done): a leaf-peeling alternative to rerooting.
