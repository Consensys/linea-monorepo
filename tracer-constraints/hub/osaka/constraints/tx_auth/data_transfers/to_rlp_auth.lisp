(module hub)


;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                             ;;
;;   X Authorization phase     ;;
;;   X.Y Value transfers       ;;
;;   X.Y.Z To RLP_AUTH         ;;
;;                             ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;


(defconst
  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___AUTH_ROW  -1
  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___ACC_ROW    0
  )

(defun  (value-transfer-to-AUTH-row)  (*  (shift  PEEK_AT_AUTHORIZATION   ROFF___VALUE_TRANSFER_TO_AUTH_ROW___AUTH_ROW )
                                          (shift  PEEK_AT_ACCOUNT         ROFF___VALUE_TRANSFER_TO_AUTH_ROW___ACC_ROW  )))


(defconstraint   authorization-phase---data-transfer---to-RLP_AUTH---nonce                                      (:guard (value-transfer-to-AUTH-row))
                 (eq!  (shift  auth/AUTHORITY_NONCE  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___AUTH_ROW )
                       (shift  account/NONCE         ROFF___VALUE_TRANSFER_TO_AUTH_ROW___ACC_ROW  )
                       ))

(defconstraint   authorization-phase---data-transfer---to-RLP_AUTH---authority-has-empty-code-or-is-delegated   (:guard (value-transfer-to-AUTH-row))
                 (if-zero   (shift  account/HAS_CODE  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___ACC_ROW)
                            ;; HAS_CODE ≡ <faux>
                            (eq!  (shift  auth/AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___AUTH_ROW)  1)
                            ;; HAS_CODE ≡ <true>
                            (eq!  (shift  auth/AUTHORITY_HAS_EMPTY_CODE_OR_IS_DELEGATED  ROFF___VALUE_TRANSFER_TO_AUTH_ROW___AUTH_ROW )
                                  (shift  account/IS_DELEGATED                           ROFF___VALUE_TRANSFER_TO_AUTH_ROW___ACC_ROW  ))
                            ))
