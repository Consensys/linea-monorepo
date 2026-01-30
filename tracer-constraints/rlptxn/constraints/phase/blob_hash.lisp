(module rlptxn)

(defun    (blob-hash-rlp-length-countdown)    cmp/AUX_1)
(defun    (blob-hash-index)                   cmp/AUX_CCC_1)
(defun    (blob-hash-first-row-of-blob-hash)  (force-bin (* (- 1 (prev TXN)) (prev DONE))))

(defconstraint blob-hash-phase-constraints-prefix (:guard IS_BLOB_HASH)
    (if-not-zero (prev TXN)
        (rlp-compound-constraint---BYTE_STRING_PREFIX-non-trivial    0
                                                                     (blob-hash-rlp-length-countdown)
                                                                     1
                                                                     1)))

(defconstraint blob-hash-phase-constraints-setting-tot-hashes (:guard IS_BLOB_HASH)
    (if-not-zero (prev TXN) (eq! blob-hash-rlp-length-countdown
                                 (* 33 (prev txn/NUMBER_OF_BLOBS)))))

(defconstraint blob-hash-phase-constraints-propagate-rlp-size (:guard IS_BLOB_HASH)
    (if-not-zero CMP
        (eq! (blob-hash-rlp-length-countdown)
             (- (prev (blob-hash-rlp-length-countdown))
                (* (prev LC) (prev cmp/LIMB_SIZE))))))


(defconstraint blob-hash-phase-constraints-each-blob-hash (:guard IS_BLOB_HASH)
    (if-not-zero (blob-hash-first-row-of-blob-hash)
        (rlp-compound-constraint---BLOBHASH   0
                                              cmp/EXO_DATA_1
                                              cmp/EXO_DATA_2)))

(defconstraint blob-hash-phase-constraints-index (:guard IS_BLOB_HASH)
        (eq!    (blob-hash-index)
                (* CMP
                   (prev CMP)
                   (+ (prev (blob-hash-index))
                      (prev DONE)))))


(defconstraint blob-hash-phase-constraints-phase-end (:guard IS_BLOB_HASH)
        (eq!    PHASE_END
                (~ (blob-hash-rlp-length-countdown))))
