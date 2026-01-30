(module hub)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;;                                          ;;;;
;;;;    X.5 Account consistency constraints   ;;;;
;;;;                                          ;;;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;


;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                       ;;
;;    X.5.4 Initialization Constraints   ;;
;;                                       ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint    account-consistency---initialization---conflation-level  (:guard   acp_FIRST_IN_CNF)
                  (begin
                    (eq!        acp_TRM_FLAG    1)
                    (vanishes!  acp_DEPLOYMENT_NUMBER)
                    (vanishes!  acp_DELEGATION_NUMBER)
                    )

(defconstraint    account-consistency---initialization---block-level       (:guard   acp_FIRST_IN_BLK)
                  (begin
                    (eq!    acp_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK    acp_DEPLOYMENT_NUMBER)
                    (eq!    acp_EXISTS_FIRST_IN_BLOCK               acp_EXISTS           )
                    ))

(defconstraint    account-consistency---initialization---transaction-level (:guard   acp_FIRST_IN_TXN)
                  (begin
                    (eq!        acp_WARMTH    acp_IS_PRECOMPILE)
                    (vanishes!  acp_DEPLOYMENT_STATUS)
                    (vanishes!  acp_MARKED_FOR_DELETION)
                    (eq! acp_HAD_CODE_INITIALLY acp_HAS_CODE)))
