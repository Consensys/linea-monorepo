package net.consensys.zkevm.domain

fun createAggregation(
  startBlockNumber: Long,
  endBlockNumber: Long = startBlockNumber + 99,
  status: Aggregation.Status = Aggregation.Status.Proven,
  aggregationCalculatorVersion: String = "1.1.1",
  batchCount: Long = 1,
  aggregationProof: ProofToFinalize = createProofToFinalize(
    firstBlockNumber = startBlockNumber,
    finalBlockNumber = endBlockNumber
  )
): Aggregation {
  return Aggregation(
    startBlockNumber = startBlockNumber.toULong(),
    endBlockNumber = endBlockNumber.toULong(),
    status = status,
    aggregationCalculatorVersion = aggregationCalculatorVersion,
    batchCount = batchCount.toULong(),
    aggregationProof = aggregationProof
  )
}

fun createAggregation(
  endBlockNumber: Long,
  parentAggregation: Aggregation,
  status: Aggregation.Status = Aggregation.Status.Proven,
  aggregationCalculatorVersion: String = "1.1.1",
  batchCount: Long = 1,
  aggregationProof: ProofToFinalize = createProofToFinalize(
    firstBlockNumber = (parentAggregation.endBlockNumber + 1UL).toLong(),
    finalBlockNumber = endBlockNumber
  )
): Aggregation {
  return Aggregation(
    startBlockNumber = parentAggregation.endBlockNumber + 1UL,
    endBlockNumber = endBlockNumber.toULong(),
    status = status,
    aggregationCalculatorVersion = aggregationCalculatorVersion,
    batchCount = batchCount.toULong(),
    aggregationProof = aggregationProof
  )
}
