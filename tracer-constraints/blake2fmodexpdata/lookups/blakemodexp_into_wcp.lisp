(defun (blake2fmodexpdata-into-wcp-oob-into-wcp-activation-flag)
  (force-bin (* (~ blake2fmodexpdata.STAMP)
                (- blake2fmodexpdata.STAMP (prev blake2fmodexpdata.STAMP)))))

;;                (defun (blake2ff-selector)
;;( * (- 1 (prev blake2fmodexpdata.IS_BLAKE_DATA)) blake2fmodexpdata.IS_BLAKE_DATA))


;;(defclookup
  ;;b-into-b
  ;;(
    ;;blake2fmodexpdata.ID
  ;;)
  ;;(blake2ff-selector)
  ;; source selector
  ;; source columns
  ;;(
    ;;blake2fmodexpdata.ID
  ;;))

(defclookup
  blake2fmodexpdata-into-wcp
  ;; target colums (in WCP)
  (
    wcp.ARG_1
    wcp.ARG_2
    wcp.RES
    wcp.INST
  )
  ;; source selector
  (blake2fmodexpdata-into-wcp-oob-into-wcp-activation-flag)
  ;; source columns
  (
    (prev blake2fmodexpdata.ID)
    blake2fmodexpdata.ID
    1
    EVM_INST_LT
  ))


