(defun (rlp-auth-into-shakiradata-for-code-hash-activation-flag) 1)

(defclookup
    (rlp-auth-into-shakiradata-for-code-hash :unchecked)
    ;; target selector
    (* shakiradata.IS_KECCAK_DATA (~ (- shakiradata.ID (prev shakiradata.ID))))
    ;; target columns
    (
        shakiradata.LIMB
        (next shakiradata.LIMB) ;; TODO: do we need also the result limbs?
        shakiradata.TOTAL_SIZE
    )
    ;; source selector
    (rlp-auth-into-shakiradata-for-code-hash-activation-flag)
    ;; source columns
    (
        keccak_for_code_hash.limb
        (next keccak_for_code_hash.limb)
        keccak_for_code_hash.total_size
    ))

;; TODO: define source selector
