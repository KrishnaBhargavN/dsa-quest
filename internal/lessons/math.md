---
title: Math for interviews
group: Advanced
order: 32
topics:
minutes: 15
summary: GCD, fast power, modular arithmetic, primes and counting. Rarely the whole problem, often the piece you need, and full of overflow traps in Java.
---

## How to recognise it

- "Return the answer modulo 10^9 + 7."
- Divisibility, gcd/lcm, fractions, slopes.
- Primes, factorisation, counting divisors.
- Counting arrangements: nCr, Catalan numbers.
- Huge exponents.

## GCD and LCM

```java
long gcd(long a, long b) { return b == 0 ? a : gcd(b, a % b); }
long lcm(long a, long b) { return a / gcd(a, b) * b; }   // divide first to avoid overflow
```

**Why Euclid works:** any common divisor of a and b also divides a − kb, hence a % b; and vice versa. So gcd(a, b) = gcd(b, a % b), and the numbers shrink at least by half every two steps: O(log min(a, b)).

Use gcd to normalise fractions and slopes: divide dy and dx by gcd(|dy|, |dx|) and fix the sign so equal slopes get equal keys.

## Fast exponentiation

```java
long power(long base, long exp, long mod) {
    long result = 1; base %= mod;
    while (exp > 0) {
        if ((exp & 1) == 1) result = result * base % mod;
        base = base * base % mod;
        exp >>= 1;
    }
    return result;
}
```

**Why it's O(log exp):** write exp in binary; base^exp is the product of base^(2^i) over the set bits i. Squaring produces each base^(2^i) in turn.

Pow(x, n) with doubles is the same loop. Convert n to `long` first: negating Integer.MIN_VALUE overflows. For negative n, use 1 / x.

## Modular arithmetic

With MOD = 1_000_000_007 (a prime):
- (a + b) % MOD, (a − b + MOD) % MOD, (a × b) % MOD with `long` operands (two values below MOD multiply to under 10^18 < 9.2 × 10^18).
- Division: multiply by the modular inverse. Since MOD is prime, inv(a) = a^(MOD − 2) mod MOD (Fermat's little theorem: a^(p−1) ≡ 1, so a · a^(p−2) ≡ 1).
- Take the mod after every operation, not just at the end.
- Java's % can return negatives; normalise with ((x % m) + m) % m.

## Primes

- Test one number: trial division up to sqrt(n). **Why sqrt:** if n = a × b with a ≤ b, then a ≤ sqrt(n), so a factor would have been found.
- All primes up to N: the sieve of Eratosthenes. For each prime p, mark p², p² + p, … as composite. Start at p² because smaller multiples were already marked by smaller primes. O(N log log N).
- Factorise many numbers: sieve the smallest prime factor spf[x], then divide repeatedly: O(log x) per number.

## Counting

- nCr mod p for many queries: precompute factorials and inverse factorials up to N; nCr = fact[n] · invFact[r] · invFact[n − r].
- Small n without mod: Pascal's triangle C[n][r] = C[n−1][r−1] + C[n−1][r], or multiply incrementally: C = C × (n − i) / (i + 1), which stays exact because each intermediate is itself a binomial.
- Grid paths right/down: C(m + n − 2, m − 1).
- Catalan numbers (valid bracket sequences, BST shapes): C(2n, n) / (n + 1).

## Big-number strings

Multiply Strings: digit i of one number times digit j of the other contributes to positions i + j and i + j + 1 of the product (from the left, product length m + n). Accumulate into an int array from the right, carry, then build the string without leading zeros. Adding numbers as strings is the same carry loop.

## Common bugs

- `int * int` overflow before the mod: cast to `long` first.
- lcm computed as a × b / gcd (overflows); divide first.
- Forgetting that `Math.pow` returns a double with rounding errors; use integer fast power.
- Negative results from `%`.

## Where it shows up

- **Pow(x, n), Multiply Strings.**
- Every "modulo 10^9 + 7" counting problem (Number of Ways to Arrive at Destination, Sum of Subarray Minimums).
- **Unique Paths:** the binomial formula.
