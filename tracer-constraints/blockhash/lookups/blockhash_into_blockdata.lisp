(defun   (blockhash-into-blockdata-selector)   blockhash.MACRO) ;; ""

(defclookup
  blockhash-into-blockdata
  ;; target columns
  (
   blockdata.RELATIVE_BLOCK_NUMBER
   blockdata.NUMBER
  )
  ;; source selector
  (blockhash-into-blockdata-selector)
  ;; source columns
  (
   blockhash.macro/REL_BLOCK
   blockhash.macro/ABS_BLOCK
   )
  )
