---
title: State-machine DP
group: Dynamic programming
order: 28
topics: DP - state machine
minutes: 11
summary: When each day you can be in one of a few modes (holding, not holding, cooling down), keep one DP value per mode and update all of them from yesterday's values.
---

## How to recognise it

- A sequence of days or steps, and at each step you are in one of a **small number of modes** with rules about switching: stock problems (holding / not holding / cooldown), "at most k transactions", painting with no two adjacent equal colours.
- The constraints are about **transitions** between consecutive steps.

## The core idea

Draw the modes as boxes and the allowed moves as arrows, each with a gain or cost. Then dp[mode] after day i is the best value of being in that mode at the end of day i, computed from day i − 1's values along the arrows into that box.

**Why it's correct:** the future only depends on which mode you're in now, not on how you got there. That's exactly what a DP state needs to capture.

## Stock with cooldown

Modes: **hold** (own a share), **sold** (just sold today), **rest** (no share, free to buy).

- hold = max(hold, rest − price)  (keep holding, or buy from rest)
- sold = hold + price             (sell what you hold)
- rest = max(rest, sold)          (keep resting, or finish cooldown)

All three must use **yesterday's** values, so compute them into temporaries first:

```java
int hold = Integer.MIN_VALUE / 2, sold = 0, rest = 0;
for (int p : prices) {
    int h = Math.max(hold, rest - p);
    int s = hold + p;
    int r = Math.max(rest, sold);
    hold = h; sold = s; rest = r;
}
return Math.max(sold, rest);
```

Start hold at a very negative number (not Integer.MIN_VALUE, which overflows when you add a price): you can't hold anything before day 0.

## Transaction fee

Two modes: cash (no share) and hold. cash = max(cash, hold + p − fee); hold = max(hold, cash − p). Using the updated cash in the hold update is fine here: selling and rebuying on the same day only loses the fee.

## At most two transactions (Stock III)

Four modes in order: after buy 1, after sell 1, after buy 2, after sell 2.

b1 = max(b1, −p); s1 = max(s1, b1 + p); b2 = max(b2, s1 − p); s2 = max(s2, b2 + p).

Updating in this order with today's values is safe: it only allows buying and selling on the same day, which gains nothing.

## At most k transactions (Stock IV)

Arrays buy[1..k] and sell[1..k]: buy[j] = max(buy[j], sell[j−1] − p), sell[j] = max(sell[j], buy[j] + p). O(n · k). If k ≥ n / 2, you can't use more than n / 2 transactions anyway, so it's the unlimited version: sum every positive day-to-day difference.

## Unlimited transactions

Sum of all max(0, p[i] − p[i−1]). **Why:** any profitable holding period splits into consecutive daily gains, and taking every positive daily gain is achievable (buy before each up-day, sell after).

## Common bugs

- Updating a mode from today's already-updated value when the rules forbid it (cooldown needs yesterday's sold).
- `Integer.MIN_VALUE + price` overflowing to a large positive.
- Returning the max over holding modes (you should end without a share).

## Where it shows up

- **Best Time to Buy and Sell Stock with Cooldown, with Transaction Fee, III and IV.**
- **Maximum Product Subarray** is a tiny state machine too (max and min ending here).
