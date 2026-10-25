---
title: Linked lists
group: Foundations
order: 9
topics: Linked list
minutes: 12
summary: Pointer surgery with a dummy head, fast and slow pointers, and in-place reversal cover nearly every linked-list problem.
---

## How to recognise it

- The input is a `ListNode`. Interviews love lists because the code is short but easy to get wrong.
- "In place", "O(1) extra space", "one pass".
- Cycles, middles, k-th from the end.

## Tool 1: the dummy head

When the head itself might change (removals, merges, reversals of the first group), create `ListNode dummy = new ListNode(0, head)` and work from it. Return `dummy.next`. This removes every "what if it's the first node" special case.

## Tool 2: reversal

```java
ListNode reverse(ListNode head) {
    ListNode prev = null, cur = head;
    while (cur != null) {
        ListNode next = cur.next;   // save before overwriting
        cur.next = prev;
        prev = cur;
        cur = next;
    }
    return prev;                    // new head
}
```

**Invariant:** `prev` is the reversed prefix, `cur` is the untouched rest. Each step moves one node across. Draw three boxes on paper the first few times; it's the most common interview whiteboard bug.

**Reverse Nodes in k-Group:** before reversing a group, walk k steps to make sure the group exists. Reverse it, then reconnect: the node before the group points to the new group head, and the old group head (now the group tail) points to the next group. Keep `groupPrev` as the node before the current group and move it to the old head after each group.

## Tool 3: fast and slow pointers

- **Middle:** slow moves 1, fast moves 2. When fast hits the end, slow is in the middle.
- **Cycle detection (Floyd):** if there's a cycle, fast eventually laps slow, because inside the cycle the gap shrinks by 1 each step, so it must hit 0.
- **Cycle start:** after they meet, reset one pointer to head and move both 1 step at a time; they meet at the cycle start.

**Why the cycle-start trick works:** let the distance from head to the cycle start be a, and the meeting point be b steps into the cycle of length c. Slow walked a + b; fast walked twice that, 2(a + b), and the extra a + b is a whole number of laps, so a + b = kc. So a = kc − b: walking a steps from the meeting point lands exactly on the cycle start, the same place a pointer from head reaches after a steps.

**Find the Duplicate Number** uses this on an array: treat i → nums[i] as `next`. Values are in 1..n and there are n + 1 indices, so following pointers from index 0 must eventually repeat; the duplicate value is the node with two incoming pointers, which is the cycle start.

## Tool 4: gaps (k-th from the end)

Move `fast` k + 1 steps ahead of `slow` (both starting at dummy), then move both until fast is null. Now slow is just before the node to remove. One pass, no length count.

## Combining tools: Reorder List

L0 → Ln → L1 → Ln−1 … is: find the middle, reverse the second half, merge the two halves alternately. Remember to cut the first half (`mid.next = null`) or you'll make a cycle.

## Copy List with Random Pointer

You need a map from old nodes to new nodes. Either a `HashMap` (two passes: create copies, then wire `next` and `random`), or the interleaving trick for O(1) space: insert each copy right after its original (A → A' → B → B'), set `copy.random = orig.random.next`, then unweave the two lists.

## LRU cache: lists as a building block

A doubly linked list gives O(1) removal of a known node and O(1) insertion at the front. Combined with a HashMap from key to node, you get O(1) "move to most recent" and "evict least recent". Use dummy head and tail nodes so you never check for null neighbours. Full design in the Design lesson.

## Merging

Merging two sorted lists is the merge step of merge sort with a dummy head. Merging k lists: a min-heap of the current heads (O(N log k)), or merging pairs of lists in rounds (also O(N log k)). Merging one by one into the result is O(N k), too slow.

## Common bugs

- Losing the rest of the list by overwriting `next` before saving it.
- Forgetting to cut a list (cycles appear, programs hang).
- Null checks: `fast != null && fast.next != null` in that order.
- Returning `head` when the head was removed or moved; return `dummy.next`.

## Where it shows up

- **Remove Nth Node From End, Reorder List, Copy List with Random Pointer, Find the Duplicate Number, LRU Cache, Reverse Nodes in k-Group, Merge k Sorted Lists.**
- **LFU Cache:** lists per frequency.
