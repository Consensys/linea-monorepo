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
  (h0h1                 :i128 :display :bytes)
  (h2h3                 :i128 :display :bytes)
  (h4h5                 :i128 :display :bytes)
  (h6h7                 :i128 :display :bytes)
  )

;; Invalid nil pointer
;; (defcall (h0h1 h2h3 h4h5 h6h7) blake2f (0 0 0 0 0 0 0 0 0 0 0 0 0 0 0))
;; works
(defcall (h0h1 h2h3 h4h5 h6h7) blake2f ( (i1 (shift LIMB 0)) (shift LIMB 1) (shift LIMB 2) (shift LIMB 3)
                                        (i1 (shift LIMB 4)) (shift LIMB 5) (shift LIMB 6) (shift LIMB 7)
                                        (i1 (shift LIMB 8)) (shift LIMB 9) (shift LIMB 10) (shift LIMB 11)
                                        (i1 (shift LIMB 12)) (shift LIMB 13) (i1 (shift LIMB 14))) )
