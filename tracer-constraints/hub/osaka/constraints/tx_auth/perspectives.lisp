(module hub)


;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                             ;;
;;   X Authorization phase     ;;
;;   X.Y Introduction          ;;
;;   X.Y Perspectives          ;;
;;                             ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;


(defconstraint    authorization-phase---perspectives---legal-perspectives                                     (:guard TX_AUTH)
                  (eq!   (+   PEEK_AT_AUTHORIZATION
                              PEEK_AT_ACCOUNT
                              PEEK_AT_TRANSACTION)
                         1))

(defconstraint    authorization-phase---perspectives---an-authorization-phase-starts-with-an-AUTH-row         (:guard TX_AUTH)
                  (if-zero   (prev   TX_AUTH)
                             (eq!    PEEK_AT_AUTHORIZATION   1)
                             ))

(defconstraint    authorization-phase---perspectives---transition---from-an-authorization-row                 (:guard TX_AUTH)
                  (if-not-zero   PEEK_AT_AUTHORIZATION
                                 (if-not-zero   auth/AUTHORITY_ECRECOVER_SUCCESS
                                                ;; Authority address successfully recovered
                                                (begin   (eq!   (next   TX_AUTH         )   1 )
                                                         (eq!   (next   PEEK_AT_ACCOUNT )   1 ))
                                                ;; Authority address recovery failure
                                                (eq!     (next   (+   PEEK_AT_AUTHORIZATION
                                                                      PEEK_AT_TRANSACTION))
                                                         1)
                                                )))

(defproperty      authorization-phase---perspectives---transition---from-an-account-row---sanity-check
                  (if-not-zero   TX_AUTH
                                 (if-not-zero   PEEK_AT_ACCOUNT
                                                (eq!    (prev   TX_AUTH)   1)
                                                )))

(defconstraint    authorization-phase---perspectives---transition---from-an-account-row---dom-sub-stamps      (:guard TX_AUTH)
                  (eq!    (prev   TX_AUTH)   1))

(defconstraint    authorization-phase---perspectives---authorization-phases-finish-on-a-TXN-row               (:guard TX_AUTH)
                  (eq!   (+   PEEK_AT_TRANSACTION
                              (next   TX_AUTH))
                         1))

(defproperty      authorization-phase---perspectives---authorization-phases-finish-on-a-TXN-row---explicit
                  (if-not-zero   TX_AUTH
                                 (if-not-zero   PEEK_AT_TRANSACTION
                                                ;; TXN[i] ≡ 1
                                                (eq!   (next  TX_AUTH)   0)
                                                ;; TXN[i] ≡ 0
                                                (eq!   (next  TX_AUTH)   1)
                                                )))

(defconstraint    authorization-phase---perspectives---transition---from-the-only-and-final-transaction-row
                  (:guard TX_AUTH)
                  (if-not-zero   PEEK_AT_TRANSACTION
                                 (begin
                                   (eq!   (next  (+  TX_SKIP  TX_INIT))                           1)
                                   (eq!   PEEK_AT_TRANSACTION                                     1)
                                   (eq!   transaction/TRANSACTION_TYPE_SUPPORTS_DELEGATION_LISTS  1)
                                   )))
