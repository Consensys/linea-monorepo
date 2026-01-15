
(module rlptxn)

(defconstraint transaction-constancies-requires-number-blob-hashes ()
  (transaction-constant NUMBER_OF_BLOBS))

  (defconstraint no-blob-tx-no-blobs ()
    (if-zero TYPE_3 (vanishes! NUMBER_OF_BLOBS)))
