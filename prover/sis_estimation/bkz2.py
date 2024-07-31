from math import log2, pi as PI, sqrt, log as ln, e as E
import math
import scipy.special
from scipy.optimize import bisect, minimize_scalar
from scipy.stats import norm

def estimate_bkz(log_q, n, log2_bound):

    log2_volume = log_q * n

    min_m = math.ceil(log2_volume / log2_bound)
    max_m = 1000000

    f = lambda m: estimate_bkz_for_m(
        m=m, 
        log2_volume=log2_volume, 
        log2_bound=log2_bound,
    )

    sec = minimize_scalar(f, bounds=(min_m, max_m), method="bounded")
    assert sec.success

    return f(sec.x)

def estimate_bkz_for_m(
    m,
    log2_volume,
    log2_bound,
):
    """
    """
    assert m*log2_bound > log2_volume

    min_log_l2_norm, max_log_l2_norm = norm_frame(
        log2_bound=log2_bound,
        m=m,
    )

    def f(log_l2_norm):
        log_t = estimate_bkz_for_m_and_l2_norm(
            m=m, 
            log2_volume=log2_volume,
            log2_l2_norm=log_l2_norm,
        )

        log_p = hypercube_l2_ball_intersection_log_prob(
            log2_bound=log2_bound, 
            m=m, 
            log2_l2_norm=log_l2_norm,
        )

        return log_t - log_p
    
    sec =  minimize_scalar(f, bounds=(min_log_l2_norm, max_log_l2_norm), method="bounded")
    assert sec.success
    return f(sec.x)


def estimate_bkz_for_m_and_l2_norm(
    m,
    log2_volume,
    log2_l2_norm,
):
    """
    Return the runtime estimate of running the BKZ attack, fixing the dimension
    of the SVP instance to be attacked. The function returns math.Inf if the
    L2 norm cannot be achieved.

    If the volume of the lattice is bigger than the L2-ball whose radius is
    equal to the target norm we employ another strategy: it is more appropriate 
    to use direct SVP attack and to examinate the probability that the output 
    vector is within the target bound, assuming a uniform distribution and GH. 
    This is also employed if the dimension m is smaller than 50 as the BKZ 
    estimates become unreliable in this regime.

        :m              : The dimension of the lattice to attack
        :log2_volume    : The volume of the lattice to attack. In the context
                            of analyzing SIS, this will be equal to the log of
                            q^n where n is the number of rows of the SIS 
                            matrix.
        :log2_l2_norm   : The log (base 2) of the target L2 norm to attack
    """

    MIN_BLOCK_SIZE = 50

    # The target root-hermite factor to achieve to match the target L2 norm
    target_rhf = get_target_log_root_hermite_factor(
        m, 
        log2_volume, 
        log2_l2_norm=log2_l2_norm,
    )

    log2_volume_target_ball = log2_volume_unit_ball(m) + m*log2_l2_norm
    if (
        m < MIN_BLOCK_SIZE
        or log2_volume_target_ball < log2_volume
        or target_rhf < bkz_log_root_hermite(k=m, n=m)
    ):
        log_t = svp_l2_oracle_cost(m) 
        log_p = log2_volume_target_ball - log2_volume
        return log_t - log_p

    # F takes a BKZ block-size parameter as input and returns a negative value
    # if the block size is too small to achieve the target rhf and positive if
    # if it can match it. Our intent is to find the block-size zeroeing this
    # function.
    f = lambda k : target_rhf - bkz_log_root_hermite(k=k, n=m)

    # block_size_0 is the minimal block size required to match the target L2 
    # norm
    block_size_0 = bisect(f=f, a=MIN_BLOCK_SIZE, b=m, disp=True)

    assert close(bkz_log_root_hermite(k=block_size_0, n=m), target_rhf)
    assert block_size_0 <= m

    return log2_runtime(k=block_size_0, n=m)


def log2_runtime(k, n):
    """
        Returns an estimation of the runtime of BKZ for a given lattice dimension and BKZ blocksize
        From APS15, Section 3.2 page 9
        https://eprint.iacr.org/2015/046.pdf
        
        k: blocksize
        n: lattice dimension
    """
    return log2(n) + log2_nb_of_bkz_rounds(k, n) + svp_l2_oracle_cost(k)


def log2_nb_of_bkz_rounds(n, k):
    """
        Returns an estimation of the log of BKZ rounds needed to achieve the 
        target root-hermite factor

        The estimation is taken from the work of LN20
        https://eprint.iacr.org/2020/1237

        k: blocksize
        n: lattice dimension
    """
    return 2 * log2(n) - 2 * log2(k) + log2(log2(n))


def get_target_log_root_hermite_factor(
    m,
    log2_volume,
    log2_l2_norm,
):
    """
        Returns the log root-hermite factor to achieve with BKZ to match the 
        required L2 norm bound.

        m: the dimension of the lattice
        log2_volume: the volume of the lattice
        log2_l2_norm: the target L2 norm to achieve
    """
    log_root_volume = log2_volume / m
    return 1/m * (log2_l2_norm - log_root_volume)


def bkz_log_root_hermite(k, n):
    """
        Returns the log-root hermite factor achievable by BKZ as a function of the block-size

        Taken from: https://eprint.iacr.org/2015/046.pdf (page 9)

        k: blocksize
        n: dimension of the lattice
    """
    return (
        log2(k / (2 * PI * E)) +
        log2(PI * k) / k
    ) / (2 * (k - 1))


def log2gamma(x):
    """
        Logarithm of the gamma function, only intended to process
        real numbers. Returns the result in base 2
    """
    return scipy.special.loggamma(x) / ln(2)


def log2_volume_unit_ball(d):
    """
        Returns the logarithm in base 2 of a unit-ball in dimension d
    """
    return d / 2 * log2(PI) - log2gamma(d/2 + 1)


def svp_l2_oracle_cost(dim):
    """
        Returns the runtime estimate of the SVP sieving algorithm of BDGL16

        https://eprint.iacr.org/2015/1128
    """
    return 0.292 * dim + 16.4

def close(a, b, prec=0.001):
    """
        Returns true if a and b are relatively close, used for assertion over
        floats.
    """
    # Rule out the case where both a and b are close to 0
    if abs(a - b) < prec:
        return True
    dist = abs(a/b - 1)
    ok = dist < prec
    if not ok:
        print('Not OK : dist {} {} {}'.format(dist, a, b))
    return ok


def norm_frame(log2_bound, m):
    return log2_bound-1, log2_bound-1 + 0.5*log2(m)

def hypercube_l2_ball_intersection_log_prob(log2_bound, m, log2_l2_norm):
    """
        Returns the probability that a normal-distributed vector V of L2 norm
        L = 2**log2_l2_norm (assuming the coordinates are independants) can fit
        in a hypercube centered at O and of side B = 2**log2_bound. m is the
        dimension of the vector (and of the hypercube).

        To clarify the assumptions on V, we say that:
            - All coordinates are normally and independantly distributed and 
            have std deviation of L/\sqrt(m)

        This follows the analysis of Estimate all the SIS parameters
    """
    # Standard deviation of one coordinate with our model
    l2_norm = (2**log2_l2_norm)
    b = 2**log2_bound
    stddev = l2_norm / sqrt(m)
    return m * log2(1 - 2*norm.cdf(-b/2, 0, stddev))
 
