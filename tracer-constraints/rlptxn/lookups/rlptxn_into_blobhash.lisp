
(defun (sel-rlptxn-to-blobhash) (force-bin (* rlptxn.IS_BLOB_HASH rlptxn.CMP)))

(defclookup
  rlptxn-into-blobhash
  ;; target columns
  (
    blobhash.USER_TXN_NUMBER
    blobhash.NUMBER_OF_BLOBS
    blobhash.HASH_INDEX
    blobhash.BLOB_VERSION_HASH
  )
  ;; source selector
  (sel-rlptxn-to-blobhash)
  ;; source columns
  (
    rlptxn.USER_TXN_NUMBER
    rlptxn.txn/NUMBER_OF_BLOBS
    rlptxn.cmp/AUX_CCC_1 ;; INDEX
    (::  rlptxn.cmp/EXO_DATA_1 rlptxn.cmp/EXO_DATA_2)
  ))
