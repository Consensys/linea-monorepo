(module hub)

;;;;;;;;;;;;;;;;;;;;;
;;                 ;;
;;   4.5 Refunds   ;;
;;                 ;;
;;;;;;;;;;;;;;;;;;;;;

;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                        ;;
;;   4.5.1 Introduction   ;;
;;                        ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;


;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                   ;;
;;   4.5.2 Gas column generalities   ;;
;;                                   ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

;; hubStamp constancies already enforced

;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                       ;;
;;   4.5.3 Constraints   ;;
;;                       ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun   (same-refund-counter   relof)   (eq!  (shift  REFUND_COUNTER_NEW  relof)
                                               (shift  REFUND_COUNTER      relof)))

(defconstraint    generalities---refunds---refund-column-constancies ()
                  (begin
                    (hub-stamp-constancy REFUND_COUNTER)
                    (hub-stamp-constancy REFUND_COUNTER_NEW)))

(defconstraint    generalities---refunds---only-USER-transactions-may-incur-refunds ()
                  (if-zero    USER
                              (begin
                                ( vanishes! REFUND_COUNTER     )
                                ( vanishes! REFUND_COUNTER_NEW ))))

(defconstraint    generalities---refunds---refunds-reset-at-transaction-boundaries ()
                  (if-not-zero   (-   TOTL_TXN_NUMBER   (prev   TOTL_TXN_NUMBER))
                                 ( vanishes! REFUND_COUNTER )))

(defconstraint    generalities---refunds---refunds-remain-constant-along-certain-transaction-processing-phases ()
                  (if-not-zero   (+   TX_SKIP
                                      TX_WARM
                                      TX_INIT
                                      TX_FINL)
                                 (same-refund-counter   0 )))

(defconstraint    generalities---refunds---what-happens-in-the-TX_AUTH-phase   (:guard   (*  TX_AUTH  PEEK_AT_AUTHORIZATION))
                  (eq!   REFUND_COUNTER_NEW
                         (+  REFUND_COUNTER
                             (auth-tuple-refund))))

(defun    (auth-tuple-refund)    (*   auth/AUTHORIZATION_TUPLE_IS_VALID
                                      (shift account/EXISTS  1)
                                      (-   GAS_CONST_PER_EMPTY_ACCOUNT   GAS_CONST_PER_AUTH_BASE_COST )))

(defun    (new-stamp-in-TX_EXEC-phase)   (*  (- HUB_STAMP (prev HUB_STAMP))
                                             TX_EXEC))

(defconstraint    generalities---refunds---what-happens-in-the-TX_EXEC-phase---initialization-constraints
                  (:guard   (new-stamp-in-TX_EXEC-phase))
                  (if-zero   (prev   TX_EXEC)
                             (eq!   REFUND_COUNTER
                                    (*  (-  1  CN_WILL_REV)
                                        (prev   REFUND_COUNTER_NEW)))))

(defconstraint    generalities---refunds---what-happens-in-the-TX_EXEC-phase---linking-constraints
                  (:guard   (new-stamp-in-TX_EXEC-phase))
                  (if-not-zero   (prev   TX_EXEC)
                                 (eq!   REFUND_COUNTER   (prev   REFUND_COUNTER_NEW))))

(defconstraint    generalities---refunds---reverting-frames-dont-accrue-refunds (:guard  CN_WILL_REV)
                  (same-refund-counter  0))

(defconstraint    generalities---refunds---non-SSTORE-opcodes-dont-accrue-refunds (:perspective stack)
                  (if-not-zero    (opcode-isnt-SSTORE)
                                  (same-refund-counter  0)))

(defun    (opcode-is-SSTORE)     (force-bin  (* stack/STO_FLAG  [stack/DEC_FLAG 2]))) ;; ""
(defun    (opcode-isnt-SSTORE)   (force-bin  (-  1  (opcode-is-SSTORE))))

;; the actual REFUND mechanics for SSTORE will be explained in the storage instruction family section
;; Note: SELFDESTRUCT doesn't accrue refunds anymore since ... Cancun or so.
