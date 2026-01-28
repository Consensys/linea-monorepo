(defclookup
    (rlp-auth-into-shakiradata-2-limbs :unchecked)
    ;; target selector
    (* shakiradata.IS_KECCAK_DATA (~ (- shakiradata.ID (prev shakiradata.ID))))
    ;; target columns
    (
        shakiradata.LIMB ;; data
        (next shakiradata.LIMB)
        (shift shakiradata.LIMB 2) ;; result
        (shift shakiradata.LIMB 3)
        shakiradata.TOTAL_SIZE
        shakiradata.IS_KECCAK_DATA
        shakiradata.IS_KECCAK_RESULT
        shakiradata.INDEX
    )
    ;; source selector
    keccak.lookup_selector
    ;; source columns
    (
        keccak.limb ;; data
        (next keccak.limb)
        (shift keccak.limb 2) ;; result
        (shift keccak.limb 3)
        keccak.total_size
        keccak.is_keccak_data
        keccak.is_keccak_result
        keccak.index
    ))
