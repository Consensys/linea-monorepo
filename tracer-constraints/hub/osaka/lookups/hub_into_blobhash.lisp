(defun (hub-into-blob-hash-activation-flag)
  (* (unexceptional-stack-row)
      hub.stack/BLOB_HASH_FLAG))

(defclookup hub-into-blob-hash
  ;; target columns
  (
   blobhash.USER_TXN_NUMBER
   blobhash.HASH_INDEX
   blobhash.BLOB_VERSION_HASH
  )
  ;; source selector
  (hub-into-blob-hash-activation-flag)
  ;; source columns
  (
   hub.USER_TXN_NUMBER
   (:: [hub.stack/STACK_ITEM_VALUE_HI 1] [hub.stack/STACK_ITEM_VALUE_LO 1]) ;; arg
   (:: [hub.stack/STACK_ITEM_VALUE_HI 4] [hub.stack/STACK_ITEM_VALUE_LO 4]) ;; result
  )
)
