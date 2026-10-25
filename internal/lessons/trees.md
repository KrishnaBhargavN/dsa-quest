---
title: Binary trees and BSTs
group: Trees and recursion
order: 11
topics: Trees
minutes: 18
summary: Most tree problems are one DFS where each call returns something about its subtree, or one BFS by levels. Decide what to pass down and what to return up.
---

## How to recognise it

- The input is a `TreeNode`, or something hierarchical.
- Level-by-level questions (views, averages, zigzag): BFS.
- Questions about paths, subtrees, heights, ancestors: DFS.
- "BST" in the statement: the sorted order is the trick.

## The two directions of information

Every DFS on a tree answers two questions:

- **What do I pass down?** Information from ancestors: the max on the path so far (Count Good Nodes), the allowed range (Validate BST), the running sum (Path Sum).
- **What do I return up?** A summary of the subtree: height, sum, "did I find p or q", the best one-sided path.

Write the function signature first: `int gain(TreeNode node)` "returns the best downward path sum starting at node". If you can say what it returns in one sentence, the recursion almost writes itself.

## Traversals

- Preorder (node, left, right): copying and serialising trees.
- Inorder (left, node, right): sorted order in a BST.
- Postorder (left, right, node): when a node needs its children's answers first (heights, DP on trees).

Iterative inorder, for when you want to stop early:

```java
Deque<TreeNode> st = new ArrayDeque<>();
TreeNode cur = root;
while (cur != null || !st.isEmpty()) {
    while (cur != null) { st.push(cur); cur = cur.left; }
    cur = st.pop();
    // visit cur  (for k-th smallest: if (--k == 0) return cur.val;)
    cur = cur.right;
}
```

## BFS by levels

```java
Deque<TreeNode> q = new ArrayDeque<>();
if (root != null) q.offer(root);
while (!q.isEmpty()) {
    int size = q.size();                 // exactly this level
    for (int i = 0; i < size; i++) {
        TreeNode node = q.poll();
        // the last i == size - 1 is the right side view
        if (node.left != null) q.offer(node.left);
        if (node.right != null) q.offer(node.right);
    }
}
```

## Validate BST: why local checks fail

Checking only `left.val < node.val < right.val` misses a deep node that violates an ancestor (a right child's left grandchild smaller than the root). Pass the allowed range down: the left subtree gets (low, node.val), the right gets (node.val, high). Use `long` bounds or `Integer` null for "no bound", because node values can be `Integer.MIN_VALUE`.

## Global answer vs returned value (Maximum Path Sum)

A path can bend at a node (left → node → right), but a path continuing **up** to the parent can only use one side. So:
- **Return** the best one-sided path: val + max(0, left, right).
- **Update the global best** with the two-sided path: val + max(0, left) + max(0, right).

Store the best in a one-element array or a field reset at the start of the public method, never a static that survives between calls (your recurring bug).

## Lowest Common Ancestor

Return the node itself if it's p or q; otherwise recurse. If both sides return non-null, the current node is where p and q split: the LCA. If only one side is non-null, pass it up.

**Why it works:** the first node (bottom-up) that sees p and q in different subtrees is the deepest common ancestor. If p is an ancestor of q, the recursion returns p before ever looking inside it, and p is indeed the LCA.

In a BST it's simpler: walk from the root; if both values are smaller go left, if both larger go right, otherwise you're at the split.

## Build from traversals

Preorder gives the root first; inorder tells you how many nodes go left of it. Keep a map from value to inorder index and a pointer into preorder; recurse on index ranges instead of copying arrays: O(n).

## Trees as graphs

For "distance K from a target", nodes need to go **up** too. Record each node's parent with one DFS, then BFS from the target treating left, right and parent as neighbours, with a visited set.

## Path sums anywhere (Path Sum III)

A downward path from an ancestor to the current node has sum (prefix at node) − (prefix at ancestor). It's Subarray Sum Equals K on each root-to-leaf path: keep a map of prefix sums along the **current path**, add on the way down, remove on the way back up (backtracking).

## Serialise and deserialise

Preorder with an explicit null marker ("#") identifies a tree uniquely. Deserialise by consuming tokens in the same preorder with a shared iterator. BFS with nulls also works (like LeetCode's own format).

## Common bugs

- Static or field state not reset between calls.
- Forgetting the null base case.
- `int` bounds in BST validation.
- Deep recursion on skewed trees with 10^5 nodes; use iterative traversals then.

## Where it shows up

- **Level Order, Right Side View, Validate BST, Kth Smallest, Count Good Nodes, Construct from Preorder and Inorder, LCA, Path Sum III, All Nodes Distance K, Maximum Path Sum, Serialize and Deserialize.**
- **House Robber III, Binary Tree Cameras:** DP on trees (next lesson in the DP group).
