# Limits for ECDATA precompiles

- line limits for ECDATA as a standard module
  - [ ] for all precompiles and their row-space in the ECDATA module
  - [ ] specify the upper limit for ECDATA
- specialized limits
  - [x] ECPAIRING 'talliers'
    - [x] PRECOMPILE_ECPAIRING_MILLER_LOOPS
    - [x] PRECOMPILE_ECPAIRING_FINAL_EXPONENTIATIONS
    - [x] PRECOMPILE_ECPAIRING_G2_MEMBERSHIP_TESTS
  - [x] ECRECOVER 'tallier'
  - [x] ECADD 'tallier'
  - [x] ECMUL 'tallier'


How large can the ECDATA trace get
- ECRECOVER: 128 * 10 = 1280
- ECADD: 16384 * 12 = 196608
- ECMUL: 32 * 10 = 320
- ECPAIRING (unexceptional, nontrivial): 16 * (4 * 12 + 2) = 800
- sum total = 199008


Number of pairs of points:
==========================
- 0 ⇐ shouldn't trigger anything in ECDATA
- 1 ⇐ strange edgecase which is permitted by the EVM
- 2 or more ⇐ canonical case

1 point:
========

- ICP fail / success
  - Failure cases: 2 * 2 * 2 - 1 = 7 ways to fail;
  - Success: 1 way to succeed
    - well-formedness of pair of points:
      - Failure: 2 ways to fail ( B not on C2 ? on C2  but not on G2 ?)
      - Success: 4 ways of succeeding:

         | small point A | ≡ ∞ | ≠ ∞ |
         | large point B | ≡ ∞ | ≠ ∞ |

         ( ★  ) case analysis:
          - [B ≡ ∞]: should trigger nothing (CS_ECPAIRING = 0, CS_G2_MEMBERSHIP = 0)
          - [B ≠ ∞] ∧ [A ≡ ∞]: should trigger nothing (CS_ECPAIRING = 0, CS_G2_MEMBERSHIP = 1)
          - [B ≠ ∞] ∧ [A ≠ ∞]: should trigger nothing (CS_ECPAIRING = 1, CS_G2_MEMBERSHIP = 0)

FAILURE_KNOWN_TO_RAM: 7 + 2  test scenarios
RAM_SUCCESS: 3 cases (or 4 is we include the distinction on [A ≡ ∞ ?] in the first case)

n points:
=========

n = number of pairs of points (n = 2, 3, 4 for the extensive tests; and maybe some larger thing to test GNARK's ability to chain the Miller loops)

- ICP failure:
- ICP success:
  - one or more of the B_k are malformed [check with 1 malformed and also with 2 or more]:
    - expect to see: (SUCCESS_BIT = 0, CS_ECPAIRING = 0, CS_G2_MEMBERSHIP = 1 @ first malformed k)
    - we have 2 ways to be malformed but we can be malformed (1) more than once (2) for different reasons
  - all of the B_k are wellformed
    - expect to see: (SUCCESS_BIT = 1)
    - if TRIVIAL_PAIRING = 1 case (all B_k's are infinity) then the arithmetization sets the RESULT to 1 + the arithmetization sets the SUCCESS_BIT = 1)
    - it TRIVIAL_PAIRING = 0
      - k by k we have the same analysis as in the 1 point case ( ★  )


FAILURE_KNOWN_TO_RAM:
- 7 + (2 * n + 4 * n(n-1)/2 + ...) = 7 + (1 + 2)^n test scenarios (the same as before + we have a choice for where it fails, how often, several failure conditions at once etc ...) 2 * n ≡ 2 ways to fail at a position 1 ≤ k ≤ n, 4 * n(n-1)/2 ways to fail at 2 positions etc ...

RAM_SUCCESS test cases we want to test:
- TRIVIAL_PAIRING = 1
  - all the B_k's are ∞, and we can freely select the A_k's to be a mixture of ∞'s and nontrivial points
  - 2 ^ n choices for (which of the A_k are ≡ ∞)
- TRIVIAL_PAIRING = 0
  - all of the pairs of points are ACCEPTABLE_PAIR_OF_POINTS
  - only some are and we require CS_G2_MEMBERSHIP_TEST
