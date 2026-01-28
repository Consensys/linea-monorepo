(module hub)

(defperspective auth

  ;; selector
  PEEK_AT_AUTH
  (
    ( TX_AUTH                                  :binary@prove)
    ( TUPLE_INDEX                              :i10)
    ( AUTHORITY_ECRECOVER_SUCCESS              :binary@prove)
    ( SENDER_IS_AUTHORITY                      :binary@prove)
    ( SENDER_IS_AUTHORITY_ACC                  :binary@prove) ;; irrelevant to RLPAUTH
    ( AUTHORITY_ADDRESS_HI                     :i32)
    ( AUTHORITY_ADDRESS_LO                     :i128)
    ( AUTHORITY_NONCE                          :i64)
    ( AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED :binary@prove)
    ( AUTHORIZATION_TUPLE_IS_VALID             :binary@prove)
    ( DELEGATION_ADDRESS_HI                    :i32)
    ( DELEGATION_ADDRESS_LO                    :i128)
    ( DELEGATION_ADDRESS_IS_ZERO               :binary@prove)
  ))

