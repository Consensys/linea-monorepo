# Vortex compilation pipeline

1. Memory-bus argument

Converts it into a log-derivative sum

2. Lookup

Convert them into log-derivative sum

3. Log-derivative sum

Same as before, group by module

4. ~~Splitter~~

Splits modules into segments of 2^22 or less. To counter 2-adicity issues. It is actually simpler to block this at the
tracer level. Enforcing low-degreeness + smallness of the traces is acceptable requirements on the arithmetization.

4. Global constraints

As before, just one quotient by segment. The evaluations points are directly shifted to save time.

5. Local constraints and openings

Just compile them into regular evaluations. Expectedly, we only evaluate the first or the last row.

6. Split-extensions

Same as before

7. MPTS: Sumcheck based

See hackmd

8. Vortex + self-recursion

See hackmd