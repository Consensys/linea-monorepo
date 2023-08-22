import math
from bkz import HEURISTIC, bkz_bit_security
from combinatorials import cpw_bit_security


class RingSisEstimation:

    ATTRIBUTE_NUM = [
        ("log2_modulus", "modulus (log2)"),
        ("degree", "degree"),
        ("log2_norm", "norm bound (log2)"),
        ("num_poly", "num_poly (log2)"),
        ("bkz_security", "bkz security"),
        ("cpw_security", "cpw security"),
        ("block_size_bytes", "block size (bytes)"),
        ("hash_size_bytes", "hash size (bytes)"),
        ("compression_ratio", "compression ratio"),
        ("num_limbs", "num limbs"),
    ]

    def __init__(self, log2_modulus, degree, log2_norm, num_poly):
        #
        # Set the inputs of the instance
        self.log2_modulus = log2_modulus
        self.degree = degree
        self.log2_norm = log2_norm
        self.num_poly = num_poly
        #
        # Run the attacks estimation cost
        self.bkz_security = bkz_bit_security(log2_modulus, degree, log2_norm, max_m=num_poly*degree) 
        # 
        self.cpw_security = cpw_bit_security(log2_modulus, degree, log2_norm, 2*num_poly*degree)
        #
        # Run the performance estimates
        self.block_size_bytes = num_poly * degree * log2_norm / 8
        self.hash_size_bytes = degree * log2_modulus / 8
        self.compression_ratio = self.block_size_bytes / self.hash_size_bytes
        self.num_limbs = math.ceil(float(log2_modulus)/log2_norm)
        
    def __str__(self):
        return ("modulus (log2) : {} | degree : {} | norm bound (log2) : {} | "
         "num_poly (log2) : {} | bkz security : {} | cpw security : {} | "
         "block size (bytes) : {} | hash size (bytes) : {} | compression ratio : {} | "
         "num limbs : {}").format(
            self.log2_modulus, self.degree, self.log2_norm, 
            self.num_poly, self.bkz_security, self.cpw_security,
            self.block_size_bytes, self.hash_size_bytes, self.compression_ratio,
            self.num_limbs,
        )
        
    def __repr__(self):
        # Fall back on __str__
        return self.__str__()

"""
    Given some of the parameters of an instance, find the smallest degree that
    fits the parameters

    log2_modulus : log2 of the norm bound
    secret_size : size of the secret, regarless of how it splits in polynomials
    log2_beta : log2 of the norm bound
"""
def find_deg(log2_modulus, log2_norm, secret_size, target_security=128, margin_bkz=0, margin_cpw=0, max_degree=1024, norm_equivalence_strategy=HEURISTIC):
    for deg in range(1, max_degree):
        # print("\tn = {}".format(deg))
        # Check for BKZ : if it's too low reiterate
        bkz_security = bkz_bit_security(log2_modulus, deg, log2_norm, max_m=secret_size, norm_equivalence_strategy=norm_equivalence_strategy)
        if bkz_security - margin_bkz < target_security:
            continue
        # Check for CPW : if security is too low, reiterate
        cpw_security = cpw_bit_security(log2_modulus, deg, log2_norm, 2*secret_size)
        if cpw_security - margin_cpw < target_security:
            continue
        #
        return RingSisEstimation(log2_modulus, deg, log2_norm, math.ceil(secret_size/deg))


    
