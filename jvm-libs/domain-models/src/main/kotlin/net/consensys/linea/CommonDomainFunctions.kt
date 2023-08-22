package net.consensys.linea

object CommonDomainFunctions {
  fun batchIntervalString(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): String {
    return "[$startBlockNumber..$endBlockNumber](${endBlockNumber - startBlockNumber + 1uL})"
  }
}
