import math
import scipy.special
from scipy.optimize import bisect, minimize_scalar

PI = math.pi
E = math.e

HEURISTIC = "heuristic"
PESSIMISTIC = "pessimistic"
MIN_BLOCK_SIZE = 50

"""
    Maps of the diverse existing oracles
    Taken from `https://estimate-all-the-lwe-ntru-schemes.github.io/docs/`
"""
SVP_ORACLES = {
    "0.265 k" : lambda k : 0.265 * k,
    "0.265 k + 16.4" : lambda k : 0.265 * k + 16.4,
    "0.2975 k" : lambda k : 0.2975 * k,
    "0.265 k + ln k" : lambda k : 0.265 * k + math.log(k),
    "0.292 k" : lambda k : 0.292 * k,
    "0.292 k + 16.4" : lambda k : 0.292 * k + 16.4,
    "0.292 k + ln k" : lambda k : 0.292 * k + math.log(k),
    "0.0935 k ln k - 0.5095 k + 8.05" : lambda k : 0.0935 * k * math.log(k) - 0.5095 * k + 8.05,
    "0.125 k ln k - 0.755 k + 2.25" : lambda k : 0.125 * k * math.log(k) - 0.755 * k + 2.25,
    "0.187 k ln k - 1.019 k + 16.1" : lambda k : 0.187 * k * math.log(k) - 0.1019 * k + 16.1,
}

"""
    Estimates the bit security of a SIS instance against 
    lattice reduction techniques

    log_q: log2 of the modulus
    log_beta: log2 of the norm bound of the SIS instance
    n: size of the SIS output
"""
def bkz_bit_security(log2_q, n, log2_beta, max_m=2**14, 
    verbose=False, oracle_estimator=SVP_ORACLES["0.292 k + 16.4"],
    norm_equivalence_strategy=HEURISTIC):

    # We need to figure out the optimal value for `m`, `f` is the partial
    # function giving the security achieved using the value of `m`
    f = lambda m : estimate_bkz_security_for_m(
            log2_q=log2_q, n=n, log2_beta=log2_beta, m=m, verbose=verbose, 
            oracle_estimator=oracle_estimator, 
            norm_equivalence_strategy=norm_equivalence_strategy)

    # Get the minimal value of m to consider as the smallest one for which the
    # lattice contains a solution
    min_m = math.ceil(log2_q/log2_beta*n)
    if min_m > max_m:
        # The maximal value of m does not allow solutions to ring-SIS to exist
        return math.inf

    # The obtained value of "min_m" correspond to a delta=1. BKZ likely will not
    # manage to solve this. Thus, we scan the values of m until we find one that
    for m in range(min_m,max_m):
        if f(m) is not math.inf:
            min_m = m
            break

    if f(max_m) is math.inf:
        # For the maximal value of m, we proceed with binary research
        l, h = min_m, max_m
        while h - l > 1:
            m = (h + l) / 2
            if f(m) is math.inf:
                h = m
            else:
                l = m
        max_m = l

    # The optimizer attempts to figure out the "optimal value" of m 
    # to use to break the instance
    sec = minimize_scalar(f, bounds=(min_m, max_m), method="bounded")

    # Assert the success of the minimizer
    assert sec.success

    return f(sec.x)

"""
    Generic memoization function
"""
def memoize(f):
    mem = {}
    def wrapped(*args, **kwargs):
        mem_key = args + ("kwd", ) + tuple(sorted(kwargs.items()))
        if mem_key in mem:
            # print("cached = {}".format(mem_key))
            return mem[mem_key]
        res = f(*args, **kwargs)
        mem[mem_key] = res
        return res
    return wrapped

"""
    Estimates the cost of a BKZ attack using the given value `m`
    It is memoized, otherwise it takes forever to run
"""
def estimate_bkz_security_for_m(
    log2_q, n, log2_beta, m, verbose=False, 
    oracle_estimator=SVP_ORACLES["0.292 k + 16.4"],
    norm_equivalence_strategy=HEURISTIC):

    # 
    # Check that the lattice instance exists
    if m < log2_q/log2_beta*n:
        return math.inf

    #
    # Just a single call to the SVP oracle is enough
    if m <= MIN_BLOCK_SIZE:
        print("calling the SVP oracle directly")
        return oracle_estimator(m)

    # 
    # Computes the log2_norm_bound for L2
    log2_l2_norm_bound = log2_equivalent_norm(log2_beta, m, norm_equivalence_strategy)
    if verbose:
        print("Equivalent (log) l2 norm", log2_l2_norm_bound, "for m = ", m)
    
    # If it overtakes q, then our estimate does not mean anything
    # => Does not mean the instance is necessarily broken, but we 
    # should reject because we failed to analyze it.
    # One can still increase q and decrease n. This should help.
    if log2_l2_norm_bound >= log2_q:
        if verbose:
            print("Reject, as the L2 larger ball contains the entire space")
        return 0 #, 36 
    #
    # Determine the root hermite factor to achieve to reach the desired
    log2_delta = log2_l2_norm_bound / m - n / (m**2) * log2_q
    if verbose:
        print("log2_delta", log2_delta, "delta", 2**log2_delta)
    #
    # Sanity check: answer should be consistent with `log2_norm_achievable_with_bkz`
    assert close(log2_l2_norm_bound, log2_norm_achievable_with_bkz(n=n, m=m, log2_delta=log2_delta, log2_q=log2_q, verbose=verbose))
    #
    # Find the BKZ block size corresponding to log_delta using a solver
    # Starting from : k = 36 the function starts to decrease (which is the only sensical behaviour).
    # Thus, we ignore what happens for "too" large root hermite factors. They are insecure anyway. 
    # It is found at the minimum of `f` which is a convex function (if `m` is large enough)
    f = lambda k : bkz_log_root_hermite(k=k, n=m) - log2_delta
    if f(m) >= 0:
        # m too small, can't get small enough vectors out of BKZ => Try with larger m
        if verbose:
            best_root_hermite = 2 ** bkz_log_root_hermite(k=m, n=m)
            print("Given the value of m, the best root hermite factor we can do should be", best_root_hermite)
        return math.inf

    # The block size should be larger than some value to not cause estimation errors
    if f(MIN_BLOCK_SIZE) > 0:
        bkz_blocksize = bisect(f=f, a=MIN_BLOCK_SIZE, b=m, disp=True)
        #
        # Sanity check: make sure that the returned blocksize, achieves the expected log_delta
        assert close(log2_delta, bkz_log_root_hermite(k=bkz_blocksize, n=m))
        #
        # Sanity check: can't accept bkz_blocksize larger than m
        assert bkz_blocksize <= m
        #
        # Returns the bit-security. Don't get confused, the lattice dimension is m, not n
        return log2_runtime(k=bkz_blocksize, n=m, oracle_estimator=oracle_estimator)

    return log2_runtime(k=MIN_BLOCK_SIZE, n=m, oracle_estimator=oracle_estimator)

"""
    Returns the logarithm in base 2 of a unit-ball in dimension d
"""
def log2_volume_unit_ball(d):
    return d / 2 * math.log2(PI) - log2gamma(d/2 + 1)


"""
    Logarithm of the gamma function, only intended to process
    real numbers. Returns the result in base 2
"""
def log2gamma(x):
    return scipy.special.loggamma(x) / math.log(2)

"""
    Returns an equivalent l2 norm to attack given an l_inf norm and
    a lattice dimension. There two strategies:

    - Taking the smallest l2-ball containing the targeted l_inf norm (pessimistic)
    - Take the l2-ball with the same volume as the target l-inf ball (heuristic)
 
    log2_beta : targeted infinite norm
    m : research space dimension
    strategy : string, either of
        "pessimistic" : Take the smallest l2-ball containing the targeted l_inf norm (pessimistic)
        "heuristic" : Take the l2-ball with the same volume as the target l-inf ball (heuristic)
"""
def log2_equivalent_norm(log2_beta, m, strategy=HEURISTIC):
    # 
    if strategy is PESSIMISTIC:
        return (log2_beta-1) + 0.5 * math.log2(m)
    if strategy is HEURISTIC:
        return log2_beta - (1/m) * log2_volume_unit_ball(m)

"""
    Returns an estimation of the runtime of BKZ for a given lattice dimension and BKZ blocksize
    From APS15, Section 3.2 page 9
    https://eprint.iacr.org/2015/046.pdf
    
    k: blocksize
    n: lattice dimension
    oracle_estimator: closure returning an estimate of the runtime of a 
    single call to the SVP oracle
"""
def log2_runtime(k, n, oracle_estimator):
    # print("n repetitions ", n, "number of solver rounds", log2_nb_of_bkz_rounds(k, n), "oracle costs", oracle_estimator(k))
    return math.log2(n) + log2_nb_of_bkz_rounds(k, n) + oracle_estimator(k)

"""
    Returns an estimation of the log of BKZ rounds needed to achieve the target root-hermite
    factor
    APS15 Section 3.2 Page 10
    https://eprint.iacr.org/2015/046.pdf

    k: blocksize
    n: lattice dimension
"""
def log2_nb_of_bkz_rounds(n, k):
    return 2 * math.log2(n) - 2 * math.log2(k) + math.log2(math.log2(n))

"""
    Returns an estimate of the L2 norm achievable by BKZ
    (For the infinity norm)
    https://cims.nyu.edu/~regev/papers/pqc.pdf
    Section 3
    log_delta: (log2) Root hermite factor

    n: size of the output ( ows of the matrix A)
    m: number of columns to use in BKZ
    log_q: (log2) of the modulus
    log2_delta: the root hermite factor

    The formula returns an estimate for the L2 norm but
    as we have ||x||_oo <= ||x||_2 <= sqrt(n) ||x||_oo
    we directly return the result to get a conservative estimate with the infinity norm.
"""
def log2_norm_achievable_with_bkz(n, m, log2_q, log2_delta, verbose=False):
    # Returns this is the log2 norm achieved
    res = m * log2_delta + n / m * log2_q
    if verbose:
        print("log norm achievable by BKZ : ", res)
    return res

"""
    Returns the log-root hermite factor achievable by BKZ as a function of the block-size
    From APS15, Section 3.2 page 9
    https://eprint.iacr.org/2015/046.pdf

    The estimation is refined (the exponent part) using the work of LN20
    https://eprint.iacr.org/2020/1237

    k: blocksize
    n: dimension of the lattice
"""
def bkz_log_root_hermite(k, n):
    # If k, is below 36, then the formula does not hold
    # print("k = {}".format(k))*
    return - 1/(2*(n**2)*k*(k-1)) * (n*(n-1) + k*(k-2)) * log2_volume_unit_ball(k)

"""
    Returns true if a and b are relatively close, utility for debugging
"""
def close(a, b, prec=0.001):
    # Rule out the case where both a and b are close to 0
    if abs(a - b) < prec:
        return True
    dist = abs(a/b - 1)
    ok = dist < prec
    if not ok:
        print('Not OK : dist {} {} {}'.format(dist, a, b))
    return ok
