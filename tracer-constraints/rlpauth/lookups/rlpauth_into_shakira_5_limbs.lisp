(defclookup
    (rlp-auth-into-shakiradata-5-limbs :unchecked)
    ;; target selector
    (* shakiradata.IS_KECCAK_DATA (~ (- shakiradata.ID (prev shakiradata.ID))))
    ;; target columns
    (
        shakiradata.ID
        shakiradata.LIMB ;; data
        (next shakiradata.LIMB)
        (shift shakiradata.LIMB 2)
        (shift shakiradata.LIMB 3)
        (shift shakiradata.LIMB 4)
        (shift shakiradata.LIMB 5) ;; result
        (shift shakiradata.LIMB 6)
        shakiradata.TOTAL_SIZE
        shakiradata.IS_KECCAK_DATA
        shakiradata.IS_KECCAK_RESULT
        shakiradata.INDEX
    )
    ;; source selector
    (- 1 (~ (- keccak.lookup_selector 5)))
    ;; source columns
    (
        keccak.id
        keccak.limb ;; data
        (next keccak.limb)
        (shift keccak.limb 2)
        (shift keccak.limb 3)
        (shift keccak.limb 4)
        (shift keccak.limb 5) ;; result
        (shift keccak.limb 6)
        keccak.total_size
        keccak.is_keccak_data
        keccak.is_keccak_result
        keccak.index
    ))
