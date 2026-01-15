(defclookup
  blockdata-into-itself
  ;; target columns
  (
   NUMBER_PARENT
   GAS_LIMIT_PARENT
   GAS_USED_PARENT
   TIMESTAMP_PARENT
   BASE_FEE_PER_GAS_PARENT
   BLOB_GAS_USED_PARENT
   EXCESS_BLOB_GAS_PARENT
  )
  ;; source selector
  IS_NOT_FIRST_BLOCK
  ;; source columns
  (
   NUMBER
   GAS_LIMIT
   GAS_USED
   TIMESTAMP
   BASE_FEE_PER_GAS
   BLOB_GAS_USED
   EXCESS_BLOB_GAS
  ))
