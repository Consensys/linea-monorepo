(module blake2fmodexpdata)

(defun (blakemodexp-first-row)
  (!= blake2fmodexpdata.STAMP (prev blake2fmodexpdata.STAMP) ))

(defun (blakemodexp-into-small-modexp-selector)
   (and! (blakemodexp-first-row)
         (== 1 blake2fmodexpdata.SMALL_MODEXP)))

(defcall

  ;; result
  (
  (::
  (shift blake2fmodexpdata.LIMB 254 )
  (shift blake2fmodexpdata.LIMB 255 )
  ))

  modexp_u256

  ;; base
  (
  (::
  (shift blake2fmodexpdata.LIMB 0   )
  (shift blake2fmodexpdata.LIMB 1   )
  )

  ;; exp
  (::
  (shift blake2fmodexpdata.LIMB 64  )
  (shift blake2fmodexpdata.LIMB 65  )
  )

  ;; modulus
  (::
  (shift blake2fmodexpdata.LIMB 128 )
  (shift blake2fmodexpdata.LIMB 129 )
  )
  )

  ;; source selector
  (blakemodexp-into-small-modexp-selector)
)
