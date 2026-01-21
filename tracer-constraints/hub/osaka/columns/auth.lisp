(module hub)

(defperspective auth

  ;; selector
  PEEK_AT_AUTH
  (
    ( TX_AUTH                                  :binary@prove)
    ( AUTHORITY_ECRECOVER_SUCCESS              :binary@prove)
    ( AUTHORITY_ADDRESS_HI                     :i32)
    ( AUTHORITY_ADDRESS_LO                     :i128)
    ( AUTHORITY_NONCE                          :i256)
    ( AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED :binary@prove)
    ( DELEGATION_ADDRESS_HI                    :i32)
    ( DELEGATION_ADDRESS_LO                    :i128)
    ( DELEGATION_ADDRESS_IS_ZERO               :binary@prove)
    ( AUTHORIZATION_TUPLE_IS_VALID             :binary@prove)
  ))

