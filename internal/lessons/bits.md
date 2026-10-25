---
title: Bit manipulation
group: Advanced
order: 31
topics: Bits and math
minutes: 11
summary: Integers are arrays of 32 bits. A handful of identities (x & (x−1), XOR cancelling, per-bit counting) solve most bit problems in O(1) space.
---

## How to recognise it

- "Without extra space", "appears once while others appear twice/three times".
- XOR in the statement, or "maximise XOR".
- Powers of two, counting set bits, subsets as masks.
- "Without using + or −".

## Operators in Java

- `&` and, `|` or, `^` xor, `~` not.
- `<<` shift left (×2), `>>` arithmetic shift right (keeps the sign), `>>>` logical shift right (fills with 0). Use `>>>` when treating an `int` as unsigned.
- Java `int` is 32-bit two's complement: −x == ~x + 1.

## Identities worth memorising

| Expression | Meaning |
| --- | --- |
| `x & (x - 1)` | x with its lowest set bit cleared |
| `x & -x` | only the lowest set bit of x |
| `(x & (x - 1)) == 0` | x is a power of two (for x > 0) |
| `a ^ a == 0`, `a ^ 0 == a` | XOR cancels pairs |
| `(x >> i) & 1` | bit i of x |
| `Integer.bitCount(x)` | number of set bits |

Set bit i with `x OR (1 << i)` (the `|` operator) and clear it with `x & ~(1 << i)`.

**Why `x & (x − 1)` clears the lowest set bit:** subtracting 1 flips the lowest set bit to 0 and every zero below it to 1. AND with the original keeps the higher bits and zeroes everything from that bit down.

## XOR tricks

- **Single Number (others twice):** XOR everything; pairs cancel, the single remains. XOR is commutative and associative, so order doesn't matter.
- **Missing number in 0..n:** XOR all indices and values.
- **Two singles (others twice):** XOR all to get a ^ b; any set bit of it (say the lowest, `x & -x`) is a bit where a and b differ; split the numbers by that bit and XOR each group.

## Counting bits modulo k (Single Number II)

When every number appears 3 times except one, count how many numbers have each bit set. For bit i, the count is 3 × (something) + (bit i of the answer). So `count % 3` is the answer's bit. O(32 n) time, O(1) space. Works for any k, and with negative numbers in Java if you set bit 31 too.

## Counting bits for 0..n

bits[i] = bits[i >> 1] + (i & 1): i has the same bits as i / 2, plus its last bit. Or bits[i] = bits[i & (i − 1)] + 1. Both O(n).

## Addition without +

a ^ b adds without carries; (a & b) << 1 is the carry. Repeat with a = a ^ b, b = carry until the carry is 0. In Java, two's complement makes this work for negatives too, because `int` overflow wraps and the carry eventually shifts out after 32 rounds.

## Reverse bits, count bits

Loop 32 times, shifting the result left and taking `n & 1`, then `n >>>= 1` (unsigned shift). For counting set bits in Java's signed int, loop `while (n != 0)` with `n &= n - 1`; `n > 0` fails for negative inputs.

## Common bugs

- `>>` instead of `>>>` for unsigned behaviour (infinite loops on negatives).
- Precedence: `a & b == 0` is `a & (b == 0)`; parenthesise.
- `1 << 31` is negative in an `int`; use `1L << 31` when you need the value.

## Where it shows up

- **Number of 1 Bits, Counting Bits, Single Number II, Sum of Two Integers.**
- **Maximum XOR of Two Numbers:** a binary trie (Tries lesson).
- **Bitmask DP:** sets as integers (Bitmask DP lesson).
