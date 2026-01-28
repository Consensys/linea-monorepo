(defclookup
    (rlp-auth-into-hub :unchecked)
    ;; target selector
    (* hub.auth/TX_AUTH hub.PEEK_AT_AUTH)
    ;; target columns
    (
        ;; TODO: lips wip
        hub.auth/AUTHORITY_ECRECOVER_SUCCESS 
        (:: hub.auth/AUTHORITY_ADDRESS_HI hub.auth/AUTHORITY_ADDRESS_LO)
        hub.auth/AUTHORITY_NONCE
        hub.auth/AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED 
        (:: hub.auth/DELEGATION_ADDRESS_HI hub.auth/DELEGATION_ADDRESS_LO)
        hub.auth/DELEGATION_ADDRESS_IS_ZERO
        hub.auth/AUTHORIZATION_TUPLE_IS_VALID
        hub.auth/SENDER_IS_AUTHORITY
        hub.auth/TUPLE_INDEX
        hub.USER_TXN_NUMBER
    )
    ;; source selector
    (* rlpauth.xtern rlpauth.authority_ecrecover_success)
    ;; source columns
    (
        rlpauth.authority_ecrecover_success ;; This is justified in RLPAUTH
        rlpauth.authority_address ;; This is justified in RLPAUTH
        rlpauth.authority_nonce ;; This is justified in the HUB
        rlpauth.authority_has_empty_code_or_is_delegated ;; This is justified in the HUB
        rlpauth.delegation_address ;; This is justified in RLPAUTH
        rlpauth.delegation_address_is_zero ;; This is justified in RLPAUTH    
        rlpauth.authorization_tuple_is_valid ;; Computed in RLPAUTH module using local computations and information from the HUB
        rlpauth.sender_is_authority ;; Computed in RLPAUTH module using local computations from TXNDATA
        rlpauth.tuple_index ;; This is justified in the HUB
        rlpauth.user_txn_number
    ))
