(defun (sel-rlptxn-to-rlpauth) (* rlptxn.IS_AUTHORIZATION_LIST (* (prev rlptxn.CMP) rlptxn.CMP )))

(defclookup
  rlptxn-to-rlpauth
  ;; target columns
  (
    rlpauth.user_txn_number
    rlpauth.tuple_index
    rlpauth.chain_id
    rlpauth.delegation_address
    rlpauth.authority_nonce
    rlpauth.y_parity
    rlpauth.r
    rlpauth.s
  )
  ;; source selector
  (sel-rlptxn-to-rlpauth)
  ;; source columns
  (
    rlptxn.USER_TXN_NUMBER
    (i10 rlptxn.cmp/AUX_CCC_1) ;; tuple_index
    (:: rlptxn.cmp/AUX_8 rlptxn.cmp/AUX_3) ;; chain_id
    (:: rlptxn.cmp/AUX_CCC_4 rlptxn.cmp/AUX_CCC_5) ;; delegation_address
    rlptxn.cmp/AUX_CCC_2 ;; authority_nonce
    (i8 rlptxn.cmp/AUX_CCC_3) ;; y_parity
    (:: rlptxn.cmp/AUX_4 rlptxn.cmp/AUX_5) ;; r
    (:: rlptxn.cmp/AUX_6 rlptxn.cmp/AUX_7) ;; s
  ))