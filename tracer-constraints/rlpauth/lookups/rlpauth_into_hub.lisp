(defclookup
    (rlp-auth-into-hub :unchecked)
    ;; target selector
    (* hub.TX_AUTH hub.PEEK_AT_AUTHORIZATION)
    ;; target columns
    (
        hub.USER_TXN_NUMBER
        hub.HUB_STAMP
        hub.auth/TUPLE_INDEX
        hub.auth/AUTHORITY_ECRECOVER_SUCCESS
        (:: hub.auth/AUTHORITY_ADDRESS_HI hub.auth/AUTHORITY_ADDRESS_LO)
        hub.auth/AUTHORITY_NONCE
        hub.auth/AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED
        (:: hub.auth/DELEGATION_ADDRESS_HI hub.auth/DELEGATION_ADDRESS_LO)
        hub.auth/DELEGATION_ADDRESS_IS_ZERO
        hub.auth/AUTHORIZATION_TUPLE_IS_VALID
        hub.auth/SENDER_IS_AUTHORITY
    )
    rlpauth.dummy_one
    ;; source columns
    (
        rlpauth.user_txn_number
        rlpauth.hub_stamp
        rlpauth.tuple_index                              ;; justified in HUB
        rlpauth.authority_ecrecover_success              ;; justified in RLPAUTH
        rlpauth.authority_address                        ;; justified in RLPAUTH
        rlpauth.authority_nonce                          ;; justified in the HUB
        rlpauth.authority_has_empty_code_or_is_delegated ;; justified in the HUB
        rlpauth.delegation_address                       ;; justified in RLPAUTH
        rlpauth.delegation_address_is_zero               ;; justified in RLPAUTH
        rlpauth.authorization_tuple_is_valid             ;; justified in RLPAUTH
        rlpauth.sender_is_authority                      ;; justified in RLPAUTH
    ))
