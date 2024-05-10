from bkz2 import svp_l2_oracle_cost, hypercube_l2_ball_intersection_log_prob, \
     norm_frame, log2_volume_unit_ball
import math
from scipy.optimize import minimize_scalar
from math import ceil

def svp_attack_via_linf(
    log2_q,
    log2_bound,
    n,
    max_m=2**14, 
):
    """
        Two-level sieving attack [Aggarwal, Mukhopadhyay]
        
        https://arxiv.org/pdf/1801.02358.pdf
    """
    m = math.ceil(log2_q/log2_bound - 1)
    dim = (m+1)*n
    # Attack using direct SVP attack
    return 0.62*dim


def svp_attack_via_l2(
    log2_q,
    log2_bound,
    n,
    max_m=2**14,
): 
    min_m = ceil(log2_q/log2_bound*n)
    f = lambda m: l2_then_prob_estimate_for_m(
        log2_q=log2_q, 
        log2_bound=log2_bound,
        n=n,
        m=m,
        )
    sec = sec = minimize_scalar(f, bounds=(min_m, max_m), method="bounded")
    assert sec.success
    return f(sec.x)


def l2_then_prob_estimate_for_m(
    log2_q,
    log2_bound,
    n,
    m,
):
    best_norm = log2_q * (n/m)
    log_t = svp_l2_oracle_cost(m)
    log_p = hypercube_l2_ball_intersection_log_prob(
        log2_bound=log2_bound,
        log2_l2_norm=best_norm,
        m=m,
        )
    return log_t - log_p
