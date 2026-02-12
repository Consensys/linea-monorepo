(module blake2fmodexpdata)

(defconstraint selector-constancies ()
  (begin
  (stamp-constancy STAMP TRIVIAL_MODEXP)
  (stamp-constancy STAMP SMALL_MODEXP)
  (stamp-constancy STAMP LARGE_MODEXP)))

(defconstraint only-one-modexp-type ()
  (eq! (+ TRIVIAL_MODEXP SMALL_MODEXP LARGE_MODEXP)
       (+ IS_MODEXP_BASE IS_MODEXP_EXPONENT IS_MODEXP_MODULUS IS_MODEXP_RESULT)))

(defconstraint trivial-modexp-are-trivial (:guard TRIVIAL_MODEXP)
  (if (== IS_MODEXP_RESULT 1) (vanishes! LIMB))
)

(defconstraint small-modexp-have-small-result ()
  (if (and! (== 1 SMALL_MODEXP)
            (== 1 IS_MODEXP_RESULT)
            (== 1 (not-last-two-least-significant-limb)))

            (vanishes! LIMB)))

(defun (modexp-input) (force-bin (+ IS_MODEXP_BASE IS_MODEXP_EXPONENT IS_MODEXP_MODULUS)))

(defun (not-last-two-least-significant-limb) (~ (* (- INDEX_MAX_MODEXP INDEX) (- INDEX_MAX_MODEXP_MO INDEX))))

(defun (last-two-least-significant-limb) (force-bin (- 1 (not-last-two-least-significant-limb))))

(defcomputedcolumn (LARGE_MODEXP_LIMB_INPUT :i1)
   (* (modexp-input)
      (not-last-two-least-significant-limb)
      (limb-is-not-zero)))

(defcomputedcolumn (LARGE_MODEXP_ACC :i8 :fwd)
  (+ (prev LARGE_MODEXP_ACC)
     LARGE_MODEXP_LIMB_INPUT))

(defun (not-last-index) (~ (- INDEX_MAX_MODEXP INDEX)))

(defun (last-index) (force-bin (- 1 (not-last-index))))

(defun (limb-is-not-zero) (~ LIMB))

(defun (limb-is-not-one) (~ (- LIMB 1)))

(defcomputedcolumn (NON_TRIVIAL_MODULUS_LIMB :i1)
      (* IS_MODEXP_MODULUS
         (force-bin (+ (* (not-last-index) (limb-is-not-zero))
                       (* (last-index) (limb-is-not-zero) (limb-is-not-one))))
      ))

(defcomputedcolumn (NON_TRIVIAL_MODULUS_ACC :i7 :fwd)
    (+ (prev NON_TRIVIAL_MODULUS_ACC) NON_TRIVIAL_MODULUS_LIMB))

(defun (last-row-of-modexp-input) (force-bin (* (modexp-input) (last-index))))

(defconstraint define-modexp-case (:guard  (last-row-of-modexp-input))
  (if (!= 0 NON_TRIVIAL_MODULUS_ACC)
        (if (!= 0 LARGE_MODEXP_ACC)
          (eq! LARGE_MODEXP 1)
          (eq! SMALL_MODEXP 1)
        )
        (eq! TRIVIAL_MODEXP 1)
  )
)
