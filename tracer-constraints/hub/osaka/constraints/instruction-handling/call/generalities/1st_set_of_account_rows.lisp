(module hub)

;;;;;;;;;;;;;;;;;;;;;;;;
;;;;;;;;;;;;;;;;;;;;;;;;
;;;;                ;;;;
;;;;    X.Y CALL    ;;;;
;;;;                ;;;;
;;;;;;;;;;;;;;;;;;;;;;;;
;;;;;;;;;;;;;;;;;;;;;;;;


;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                         ;;
;;    X.Y.Z.3 First set of account rows    ;;
;;                                         ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun    (call-instruction---summon-both-account-rows-once-or-more)    (*    PEEK_AT_SCENARIO
                                                                              (scenario-shorthand---CALL---sum)
                                                                              (+    (call-instruction---STACK-oogx)    (scenario-shorthand---CALL---unexceptional))))

;;------------------------;;
;;   caller account-row   ;;
;;------------------------;;



(defconstraint    call-instruction---1st-caller-account-operation    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    (eq!    (shift    account/ADDRESS_HI    CALL_1st_caller_account_row___row_offset)    (call-instruction---current-address-hi))
                    (eq!    (shift    account/ADDRESS_LO    CALL_1st_caller_account_row___row_offset)    (call-instruction---current-address-lo))
                    ;; balance done below
                    (account-same-nonce                              CALL_1st_caller_account_row___row_offset)
                    (account-same-code                               CALL_1st_caller_account_row___row_offset)
                    (account-dont-check-for-delegation               CALL_1st_caller_account_row___row_offset)
                    (account-same-warmth                             CALL_1st_caller_account_row___row_offset)
                    (account-same-deployment-number-and-status       CALL_1st_caller_account_row___row_offset)
                    (account-same-marked-for-deletion                CALL_1st_caller_account_row___row_offset)
                    (account-dont-trigger-ROM_LEX                    CALL_1st_caller_account_row___row_offset)
                    (vanishes!    (shift    account/ROMLEX_FLAG      CALL_1st_caller_account_row___row_offset))
                    (vanishes!    (shift    account/TRM_FLAG         CALL_1st_caller_account_row___row_offset))
                    (vanishes!    (shift    account/RLPADDR_FLAG     CALL_1st_caller_account_row___row_offset))
                    (DOM-SUB-stamps---standard                       CALL_1st_caller_account_row___row_offset
                                                                     CALL_1st_caller_account_row___row_offset)
                    ))

(defconstraint    call-instruction---1st-caller-account-operation---balance-update    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    (if-not-zero    (scenario-shorthand---CALL---balance-update-not-required)
                                    (account-same-balance            CALL_1st_caller_account_row___row_offset))
                    (if-not-zero    (scenario-shorthand---CALL---balance-update-required)
                                    (account-decrement-balance-by    CALL_1st_caller_account_row___row_offset
                                                                     (* (call-instruction---is-CALL) (call-instruction---STACK-value-lo))))))


;;------------------------;;
;;   callee account-row   ;;
;;------------------------;;



(defconstraint    call-instruction---1st-callee-account-operation    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    ;; account/ADDRESS_HI set implicitly by account-trim-address
                    (debug (eq!   (shift    account/TRM_RAW_ADDRESS_HI     CALL_1st_callee_account_row___row_offset)    (call-instruction---STACK-raw-callee-address-hi))) ;; set implicitly by account-trim-address
                    (debug (eq!   (shift    account/ADDRESS_LO             CALL_1st_callee_account_row___row_offset)    (call-instruction---STACK-raw-callee-address-lo))) ;; set implicitly by account-trim-address
                    ;; balance done below
                    (account-same-nonce                                    CALL_1st_callee_account_row___row_offset)
                    (account-same-code                                     CALL_1st_callee_account_row___row_offset)
                    (account-check-for-delegation-if-account-has-code      CALL_1st_callee_account_row___row_offset)
                    ;; warmth done below
                    (account-same-deployment-number-and-status             CALL_1st_callee_account_row___row_offset)
                    (account-same-marked-for-deletion                      CALL_1st_callee_account_row___row_offset)
                    (account-dont-trigger-ROM_LEX                          CALL_1st_callee_account_row___row_offset)
                    (account-trim-address                                  CALL_1st_callee_account_row___row_offset          ;; row offset
                                                                           (call-instruction---STACK-raw-callee-address-hi)  ;; high part of raw, potentially untrimmed address
                                                                           (call-instruction---STACK-raw-callee-address-lo)  ;; low  part of raw, potentially untrimmed address
                                                                           )
                    (debug (eq!   (shift    account/TRM_FLAG               CALL_1st_callee_account_row___row_offset)    1)) ;; set implicitly by account-trim-address
                    (vanishes!    (shift    account/RLPADDR_FLAG           CALL_1st_callee_account_row___row_offset))
                    (DOM-SUB-stamps---standard                             CALL_1st_callee_account_row___row_offset
                                                                           CALL_1st_callee_account_row___row_offset)
                    ))


(defconstraint    call-instruction---1st-callee-account-operation---balance-update    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    (if-not-zero    (scenario-shorthand---CALL---balance-update-not-required)
                                    (account-same-balance            CALL_1st_callee_account_row___row_offset))
                    (if-not-zero    (scenario-shorthand---CALL---balance-update-required)
                                    (account-increment-balance-by    CALL_1st_callee_account_row___row_offset
                                                                     (* (call-instruction---is-CALL) (call-instruction---STACK-value-lo))))))

(defconstraint    call-instruction---1st-callee-account-operation---warmth-update    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    (if-not-zero    (scenario-shorthand---CALL---callee-warmth-update-not-required)
                                    (account-same-warmth             CALL_1st_callee_account_row___row_offset))
                    (if-not-zero    (scenario-shorthand---CALL---callee-warmth-update-required)
                                    (account-turn-on-warmth          CALL_1st_callee_account_row___row_offset))))


;;------------------------------------;;
;;   delegate or callee account-row   ;;
;;------------------------------------;;



(defconstraint    call-instruction---1st-delegt-account-operation    (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  (begin
                    (eq!   (shift    account/ADDRESS_HI                    CALL_1st_delegt_account_row___row_offset)    (call-instruction---explicit-delegate-or-callee-address-hi))
                    (eq!   (shift    account/ADDRESS_LO                    CALL_1st_delegt_account_row___row_offset)    (call-instruction---explicit-delegate-or-callee-address-lo))
                    (account-same-balance                                  CALL_1st_delegt_account_row___row_offset)
                    (account-same-nonce                                    CALL_1st_delegt_account_row___row_offset)
                    (account-same-code                                     CALL_1st_delegt_account_row___row_offset)
                    (account-check-for-delegation-if-account-has-code      CALL_1st_delegt_account_row___row_offset)
                    ;; warmth set below
                    (account-same-deployment-number-and-status             CALL_1st_delegt_account_row___row_offset)
                    (account-same-marked-for-deletion                      CALL_1st_delegt_account_row___row_offset)
                    (account-conditionally-trigger-ROM_LEX                 CALL_1st_delegt_account_row___row_offset    (call-instruction---trigger_ROMLEX))
                    (account-trim-address                                  CALL_1st_delegt_account_row___row_offset                     ;; row offset
                                                                           (call-instruction---explicit-delegate-or-callee-address-hi)  ;; hi part of either delegate or callee address
                                                                           (call-instruction---explicit-delegate-or-callee-address-lo)  ;; lo part of either delegate or callee address
                                                                           )
                    (debug (eq!   (shift    account/TRM_FLAG               CALL_1st_delegt_account_row___row_offset)    1)) ;; set implicitly by account-trim-address
                    (vanishes!    (shift    account/RLPADDR_FLAG           CALL_1st_delegt_account_row___row_offset))
                    (DOM-SUB-stamps---standard                             CALL_1st_delegt_account_row___row_offset
                                                                           CALL_1st_delegt_account_row___row_offset)
                    ))


(defun   (call-instruction---explicit-delegate-or-callee-address-hi)  (if-zero   (call-instruction---callee-is-delegated        )
                                                                                 (call-instruction---callee-address-hi          )  ;; callee isn't delegated ≡ <true>
                                                                                 (call-instruction---callee-delegate-address-hi )  ;; callee is    delegated ≡ <true>
                                                                                 ))
(defun   (call-instruction---explicit-delegate-or-callee-address-lo)  (if-zero   (call-instruction---callee-is-delegated        )
                                                                                 (call-instruction---callee-address-lo          )  ;; callee isn't delegated ≡ <true>
                                                                                 (call-instruction---callee-delegate-address-lo )  ;; callee is    delegated ≡ <true>
                                                                                 ))

(defconstraint    call-instruction---1st-delegt-account-operation---updating-warmth
                  (:guard (call-instruction---summon-both-account-rows-once-or-more))
                  ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                  (begin
                    (if-not-zero   (scenario-shorthand---CALL---callee-warmth-update-not-required)   (account-same-warmth      CALL_1st_delegt_account_row___row_offset))
                    (if-not-zero   (scenario-shorthand---CALL---callee-warmth-update-required)       (account-turn-on-warmth   CALL_1st_delegt_account_row___row_offset))
                    ))
