package net.consensys.linea

import org.apache.tuweni.bytes.Bytes32

data class BlockNumberAndHash(
  val number: ULong,
  val hash: Bytes32
)
