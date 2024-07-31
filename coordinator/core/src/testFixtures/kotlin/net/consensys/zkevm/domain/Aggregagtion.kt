package net.consensys.zkevm.domain

fun createAggregation(
  startBlockNumber: Long? = null,
  endBlockNumber: Long? = null,
  status: Aggregation.Status = Aggregation.Status.Proven,
  batchCount: Long = 1,
  aggregationProof: ProofToFinalize? = null
): Aggregation {
  require(
    (startBlockNumber != null && endBlockNumber != null) ||
      aggregationProof != null
  ) { "Either aggregationProof or startBlockNumber, endBlockNumber must be provided" }
  val _aggregationProof = aggregationProof ?: createProofToFinalize(
    firstBlockNumber = startBlockNumber!!,
    finalBlockNumber = endBlockNumber!!
  )

  return Aggregation(
    startBlockNumber = _aggregationProof.firstBlockNumber.toULong(),
    endBlockNumber = _aggregationProof.finalBlockNumber.toULong(),
    status = status,
    batchCount = batchCount.toULong(),
    aggregationProof = _aggregationProof
  )
}

fun createAggregation(
  endBlockNumber: Long,
  parentAggregation: Aggregation,
  status: Aggregation.Status = Aggregation.Status.Proven,
  batchCount: Long = 1,
  aggregationProof: ProofToFinalize = createProofToFinalize(
    firstBlockNumber = (parentAggregation.endBlockNumber + 1UL).toLong(),
    finalBlockNumber = endBlockNumber
  )
): Aggregation {
  return createAggregation(
    status = status,
    batchCount = batchCount,
    aggregationProof = aggregationProof
  )
}
