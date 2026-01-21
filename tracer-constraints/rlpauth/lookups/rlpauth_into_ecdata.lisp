(defun (rlp-auth-into-ecdata-activation-flag) 1)

(defclookup
    (rlp-auth-into-ecdata :unchecked)
    ;; target selector
    ;; ...
    ;; target columns
    (
        ecrecover.limb
        ecrecover.total_size
        ecrecover.success_bit
    )
    ;; source selector
    (rlp-auth-into-ecdata-activation-flag)
    ;; source columns
    (
        ecdata.LIMB
        ecdata.TOTAL_SIZE
        ecdata.SUCCESS_BIT    
    ))
