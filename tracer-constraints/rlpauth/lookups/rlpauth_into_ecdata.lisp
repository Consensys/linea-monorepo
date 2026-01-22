(defun (rlp-auth-into-ecdata-activation-flag) 1)

(defclookup
    (rlp-auth-into-ecdata :unchecked)
    ;; target selector
    (* ecdata.IS_ECRECOVER_DATA (~ (- ecdata.ID (prev ecdata.ID))))
    ;; target columns
    (
        ecdata.LIMB
        (next ecdata.LIMB)
        (shift ecdata.LIMB 2)
        (shift ecdata.LIMB 3)
        (shift ecdata.LIMB 4)
        (shift ecdata.LIMB 5)
        (shift ecdata.LIMB 6)
        (shift ecdata.LIMB 7)
        (shift ecdata.LIMB 8)
        (shift ecdata.LIMB 9)
        ecdata.TOTAL_SIZE
        ecdata.SUCCESS_BIT  
    )
    ;; source selector
    (rlp-auth-into-ecdata-activation-flag)
    ;; source columns
    (
        ecrecover.limb
        (next ecrecover.limb)
        (shift ecrecover.limb 2)
        (shift ecrecover.limb 3)
        (shift ecrecover.limb 4)
        (shift ecrecover.limb 5)
        (shift ecrecover.limb 6)
        (shift ecrecover.limb 7)
        (shift ecrecover.limb 8)
        (shift ecrecover.limb 9)
        ecrecover.total_size
        ecrecover.success_bit
    ))

;; TODO: define source selector
