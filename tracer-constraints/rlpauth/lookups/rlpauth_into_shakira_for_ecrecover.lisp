(defun (rlp-auth-into-shakiradata-for-ecrecover-activation-flag) 1)

(defclookup
    (rlp-auth-into-shakiradata-for-ecrecover :unchecked)
    ;; target selector
    (* shakiradata.IS_KECCAK_DATA (~ (- shakiradata.ID (prev shakiradata.ID))))
    ;; target columns
    (
        shakiradata.LIMB
        (next shakiradata.LIMB)
        (shift shakiradata.LIMB 2)
        (shift shakiradata.LIMB 3) 
        (shift shakiradata.LIMB 4)
        shakiradata.TOTAL_SIZE
    )
    ;; source selector
    (rlp-auth-into-shakiradata-for-ecrecover-activation-flag)
    ;; source columns
    (
        keccak_for_ecrecover.limb ;; data
        (next keccak_for_ecrecover.limb)
        (shift keccak_for_ecrecover.limb 2)
        (shift keccak_for_ecrecover.limb 3) ;; result
        (shift keccak_for_ecrecover.limb 4)
        keccak_for_ecrecover.total_size
    ))
