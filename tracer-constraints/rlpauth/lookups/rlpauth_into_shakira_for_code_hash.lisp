(defun (rlp-auth-into-shakiradata-for-code-hash-activation-flag) 1)

(defclookup
    (rlp-auth-into-shakiradata-for-code-hash :unchecked)
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
    (rlp-auth-into-shakiradata-for-code-hash-activation-flag)
    ;; source columns
    (
        keccak_for_code_hash.limb
        (next keccak_for_code_hash.limb)
        (shift keccak_for_code_hash.limb 2)
        (shift keccak_for_code_hash.limb 3)
        (shift keccak_for_code_hash.limb 4)
        keccak_for_code_hash.total_size
    ))

;; TODO: define source selector
