(defun (rlp-auth-into-ecdata-activation-flag) 1)

(defclookup
    (rlp-auth-into-ecdata :unchecked)
    ;; target selector
    (* ecdata.IS_ECRECOVER_DATA (~ (- ecdata.ID (prev ecdata.ID))))
    ;; target columns
    (
        ecdata.LIMB
        ecdata.TOTAL_SIZE
        ecdata.SUCCESS_BIT  
    )
    ;; source selector
    (rlp-auth-into-ecdata-activation-flag)
    ;; source columns
    (
        ecrecover.limb
        ecrecover.total_size
        ecrecover.success_bit
    ))

;; TODO: define source selector
