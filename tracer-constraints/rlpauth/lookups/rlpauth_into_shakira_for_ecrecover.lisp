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
        (shift shakiradata.LIMB 5)
        (shift shakiradata.LIMB 6)
        (shift shakiradata.LIMB 7)
        (shift shakiradata.LIMB 8)
        (shift shakiradata.LIMB 9)
        (shift shakiradata.LIMB 10)
        shakiradata.TOTAL_SIZE
    )
    ;; source selector
    (rlp-auth-into-shakiradata-for-ecrecover-activation-flag)
    ;; source columns
    (
        keccak_for_ecrecover.limb
        (next keccak_for_ecrecover.limb)
        (shift keccak_for_ecrecover.limb 2)
        (shift keccak_for_ecrecover.limb 3)
        (shift keccak_for_ecrecover.limb 4)
        (shift keccak_for_ecrecover.limb 5)
        (shift keccak_for_ecrecover.limb 6)
        (shift keccak_for_ecrecover.limb 7)
        (shift keccak_for_ecrecover.limb 8)
        (shift keccak_for_ecrecover.limb 9)
        (shift keccak_for_ecrecover.limb 10)
        keccak_for_ecrecover.total_size
    ))

;; TODO: define source selector
