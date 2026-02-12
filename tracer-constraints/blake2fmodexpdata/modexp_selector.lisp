(module blake2fmodexpdata)

(defconstraint selector-constancies ()
  (begin
  (stamp-constancy STAMP TRIVIAL_MODEXP)
  (stamp-constancy STAMP SMALL_MODEXP)
  (stamp-constancy STAMP LARGE_MODEXP)))

(defconstraint only-one-modexp-type ()
  (eq! (+ TRIVIAL_MODEXP SMALL_MODEXP LARGE_MODEXP)
       (+ IS_MODEXP_BASE IS_MODEXP_EXPONENT IS_MODEXP_MODULUS IS_MODEXP_RESULT)))

(defconstraint trivial-modexp-are-trivial ()
  (if (and! (== IS_MODEXP_RESULT 1)
            (== 1 TRIVIAL_MODEXP))
  (vanishes! LIMB))
)

(defconstraint small-modexp-have-small-result ()
  (if (and! (== 1 SMALL_MODEXP)
            (== 1 (not-last-two-least-significant-limb-of-result)))

            (vanishes! LIMB)))

(defun (modexp-input) (force-bin (+ IS_MODEXP_BASE IS_MODEXP_EXPONENT IS_MODEXP_MODULUS)))

(defun (not-last-two-least-significant-limb-of-result)
    (* IS_MODEXP_RESULT   (shift IS_MODEXP_RESULT   2)))

(defun (not-last-two-least-significant-limb-of-input)
    (force-bin (+ (* IS_MODEXP_BASE     (shift IS_MODEXP_BASE     2))
                  (* IS_MODEXP_EXPONENT (shift IS_MODEXP_EXPONENT 2))
                  (* IS_MODEXP_MODULUS  (shift IS_MODEXP_MODULUS  2))))
    )

(defcomputedcolumn (LARGE_MODEXP_LIMB_INPUT :i1)
   (* (not-last-two-least-significant-limb-of-input)
      (limb-is-not-zero)))

(defcomputedcolumn (LARGE_MODEXP_ACC :i8 :fwd)
  (+ (prev LARGE_MODEXP_ACC)
     LARGE_MODEXP_LIMB_INPUT))

(defun (last-index-of-modulus)
    (force-bin (* IS_MODEXP_MODULUS
                  (- 1 (next IS_MODEXP_MODULUS)))))

(defun (limb-is-not-zero) (~ LIMB))

(defun (limb-is-not-one) (~ (- LIMB 1)))

(defcomputedcolumn (NON_TRIVIAL_MODULUS_LIMB :i1)
        (if (!= 0 (limb-is-not-zero))
            ;; case LIMB != 0
	          (if (== 1 (last-index-of-modulus))
	                ;; case LAST_INDEX: true if LIMB !=1, else true
	                (* (limb-is-not-one) IS_MODEXP_MODULUS)
	                ;; always true
	                IS_MODEXP_MODULUS)
	          ;; case LIMB == 0
	          0)
  )

(defcomputedcolumn (NON_TRIVIAL_MODULUS_ACC :i7 :fwd)
       (+ (prev NON_TRIVIAL_MODULUS_ACC) NON_TRIVIAL_MODULUS_LIMB))

(defun (last-row-of-modexp-input) (force-bin (* IS_MODEXP_MODULUS (- 1 (next IS_MODEXP_MODULUS)))))

(defconstraint define-modexp-case (:guard  (last-row-of-modexp-input))
  (if (!= 0 NON_TRIVIAL_MODULUS_ACC)
        (if (!= 0 LARGE_MODEXP_ACC)
          (eq! LARGE_MODEXP 1)
          (eq! SMALL_MODEXP 1)
        )
        (eq! TRIVIAL_MODEXP 1)
  )
)
