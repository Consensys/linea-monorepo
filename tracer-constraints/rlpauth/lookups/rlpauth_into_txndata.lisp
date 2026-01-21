(defun (rlp-auth-into-txndata-activation-flag) (* rlpauth.macro rlpauth.authority_ecrecover_success))

(defclookup
    (rlp-auth-into-txndata :unchecked)
    ;; target selector
    (* txndata.USER txndata.HUB (prev txndata.rlp/TYPE_4))
    ;; target columns
    (
        txndata.BLK_NUMBER
        txndata.USER_TXN_NUMBER        
        (:: txndata.hub/FROM_ADDRESS_HI txndata.hub/FROM_ADDRESS_LO)
        ;; txndata.AUTHORITY_IS_SENDER_TOT ;; TODO
    )
    ;; source selector
    (rlp-auth-into-txndata-activation-flag)
    ;; source columns
    (
        rlpauth.blk_number
        rlpauth.user_txn_number
        rlpauth.txn_from_address
        ;; rlpauth.authority_is_sender_tot          
    ))
