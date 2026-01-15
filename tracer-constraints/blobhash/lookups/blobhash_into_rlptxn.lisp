(defun (sel-src-blobhash-into-rlptxn) (force-bin (- 1 blobhash.OOI)))
(defun (sel-trg-blobhash-into-rlptxn) (sel-rlptxn-to-blobhash))

(defclookup
  (blobhash-into-rlptxn :unchecked)
  ;; target selector
  (sel-trg-blobhash-into-rlptxn)
  ;; target columns
  (
    rlptxn.USER_TXN_NUMBER
    rlptxn.cmp/AUX_CCC_1 ;; TOT number of hashes
    rlptxn.cmp/AUX_CCC_2 ;; INDEX
    (::  rlptxn.cmp/EXO_DATA_1 rlptxn.cmp/EXO_DATA_2)
  )
  ;; source selector
  (sel-src-blobhash-into-rlptxn)
  ;; source columns
  (
    blobhash.USER_TXN_NUMBER
    blobhash.TOT_NUMBER_OF_HASHES
    blobhash.HASH_INDEX
    blobhash.BLOB_VERSION_HASH
  ))
