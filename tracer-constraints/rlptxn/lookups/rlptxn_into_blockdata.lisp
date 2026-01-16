(defun (sel-rlptxn-to-blockdata) (force-bin (* rlptxn.TXN rlptxn.REPLAY_PROTECTION)))

(defclookup
  rlptxn-into-blockdata
  ;; target columns
  (
    blockdata.CHAINID
  )
  ;; source selector
  (sel-rlptxn-to-blockdata)
  ;; source columns
  (
    rlptxn.txn/CHAIN_ID
  ))
