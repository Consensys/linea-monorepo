# Hierarchical aggregation

This short document describes the hierarchical aggregation scheme. The scheme
allows aggregating the segment subproofs 2-by-2 in any order to reduce the 
latency as much as possible.

Crucially, the scheme relies on an LtHash, an order-independant hash function. 
The proofs are recursively verified as we do with the all-at-once conglomeration
, but the main change is that the scheme allow recursively aggregating the 
conglomeration proofs. The list of the "acceptable" verification keys of the 
aggregated proofs can no longer be fixed when defining the hierarchical 
conglomeration circuit and must be defined as public input which is ultimately
enforced by the verifier.

A second aspect of the hierarchical conglomeration is how the public inputs are
managed during the aggregation. Crucially, the conglomeration must guarantee the
following properties.

- All the segment proofs are from a list of whitelisted circuits
- At the end of the process, all the segments have been verified
- All the segments have the same view of the randomness generation and the
    randomness has been correctly generated.
- The logderivative sum, grand product etc.. are holistically valid
- The "from-prev-to-next-segments" checks (HornerN0Hash and GlobalSendReceive) 
    are verified
- The LPP commitments are consistent between GL and LPP of the same segments and 
    the pair should have verification keys for the same module
- The functional inputs are propagated up through the aggregation steps

