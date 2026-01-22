(defun (rlp-auth-into-shakiradata-activation-flag) 1)

(defclookup
    (rlp-auth-into-shakiradata :unchecked)
    ;; target selector
    (* shakiradata.IS_KECCAK_DATA (~ (- shakiradata.ID (prev shakiradata.ID))))
    ;; target columns
    (
        shakiradata.LIMB
        shakiradata.TOTAL_SIZE
    )
    ;; source selector
    (rlp-auth-into-shakiradata-activation-flag)
    ;; source columns
    (
        keccak_for_ecrecover.limb
        keccak_for_ecrecover.total_size
    ))

;; TODO: define source selector
