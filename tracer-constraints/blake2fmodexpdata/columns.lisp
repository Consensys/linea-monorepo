(module blake2fmodexpdata)

(defcolumns
  (STAMP                :i16)
  (ID                   :i32)
  (PHASE                :byte)
  (INDEX                :byte :display :dec)
  (INDEX_MAX            :byte :display :dec)
  (LIMB                 :i128 :display :bytes)
  (IS_MODEXP_BASE       :binary@prove)
  (IS_MODEXP_EXPONENT   :binary@prove)
  (IS_MODEXP_MODULUS    :binary@prove)
  (IS_MODEXP_RESULT     :binary@prove)
  (IS_BLAKE_DATA        :binary@prove)
  (IS_BLAKE_PARAMS      :binary@prove)
  (IS_BLAKE_RESULT      :binary@prove)
  (h0h1_be              :i128 :display :bytes)
  (h2h3_be              :i128 :display :bytes)
  (h4h5_be              :i128 :display :bytes)
  (h6h7_be              :i128 :display :bytes)
  )

(defun (blake2f-selector)
 ( * (- 1 (prev blake2fmodexpdata.IS_BLAKE_DATA)) blake2fmodexpdata.IS_BLAKE_DATA))

(defun (tmp-selector)
  (if (== blake2fmodexpdata.INDEX 0)
      (if (== blake2fmodexpdata.STAMP 1)
          (if (== blake2fmodexpdata.PHASE 0x5) 1 0) 0 ) 0 ))

;; Invalid nil pointer
;; (defcall (h0h1 h2h3 h4h5 h6h7) blake2f (0 0 0 0 0 0 0 0 0 0 0 0 0 0 0))

(defcall (h0h1_be h2h3_be h4h5_be h6h7_be) blake2f ( (i64 (shift LIMB 0)) (shift LIMB 1) (shift LIMB 2) (shift LIMB 3)
                                        (i1 (shift LIMB 4)) (shift LIMB 5) (shift LIMB 6) (shift LIMB 7)
                                        (i1 (shift LIMB 8)) (shift LIMB 9) (shift LIMB 10) (shift LIMB 11)
                                        (i1 (shift LIMB 12)) (shift LIMB 13) (i1 (shift LIMB 14))) (== 1 (tmp-selector)))
