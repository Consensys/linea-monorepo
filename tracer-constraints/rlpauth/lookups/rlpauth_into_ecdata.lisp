(defun (rlp-auth-into-ecdata-activation-flag) 1)

(deflookup
    (rlp-auth-into-ecdata :unchecked)
    ;; target columns
    (
        ecdata.LIMB
        ecdata.TOTAL_SIZE
        ecdata.SUCCESS_BIT  
    )
    ;; source columns
    (
        ecrecover.limb
        ecrecover.total_size
        ecrecover.success_bit
    ))

;; TODO: define selectors, add index?