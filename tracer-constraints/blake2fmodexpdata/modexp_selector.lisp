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
  (if-not-zero IS_MODEXP_RESULT (vanishes! LIMB))
)

(defconstraint small-modexp-have-small-result (:guard SMALL_MODEXP)
  (if (and! (== 1 IS_MODEXP_RESULT)
            (== 1 (not-least-two-least-significant-limb)))

            (vanishes! LIMB)))

(defun (modexp-input) (force-bin (+ IS_MODEXP_BASE IS_MODEXP_EXPONENT IS_MODEXP_MODULUS)))

(defun (not-least-two-least-significant-limb) (~ (* (- INDEX_MAX_MODEXP INDEX) (- INDEX_MAX_MODEXP_MO INDEX))))

(defcomputedcolumn (LARGE_MODEXP_LIMB_INPUT :i1)
   (* (not-least-two-least-significant-limb)
      (modexp-input)
      (~ LIMB)))

(defcomputedcolumn (LARGE_MODEXP_ACC :i8 :fwd)
    (* (modexp-input) (+ LARGE_MODEXP_ACC
                         LARGE_MODEXP_LIMB_INPUT)))

(defun (last-row-of-modexp-input) (force-bin (* (modexp-input) (- 1 (modexp-input)))))

(defconstraint define-modexp-case (:guard  (last-row-of-modexp-input))
  (if-not-zero LARGE_MODEXP_ACC
                                ;; if ACC !=0, then at least of the input is > u32, so LARGE_MODEXP
                                (eq! LARGE_MODEXP 1)
                                ;; need to differentiate between small and trivial (ie modulus == 0 or 1)) modexp
                                ;; prev LIMB (ie hi part of the i32 input) check
                                (if-not-zero (prev LIMB) (eq! SMALL_MODEXP 1)
                                                         (if-not (or! (== LIMB 0) (== LIMB 1))
                                                              (eq! SMALL_MODEXP 1)
                                                              (eq! TRIVIAL_MODEXP 1) ))))

