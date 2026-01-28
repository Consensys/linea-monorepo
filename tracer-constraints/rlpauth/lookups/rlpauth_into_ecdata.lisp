(defun (rlp-auth-into-ecdata-activation-flag) 1)

(defclookup
    (rlp-auth-into-ecdata :unchecked)
    ;; target selector
    (* ecdata.IS_ECRECOVER_DATA (~ (- ecdata.ID (prev ecdata.ID)))) ;; alternatively (* ecdata.IS_ECRECOVER_DATA (- 1 (prev ecdata.IS_ECRECOVER_DATA)))
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
        ecrecover.total_size
        ecrecover.success_bit
    ))

;; TODO: define source selector
