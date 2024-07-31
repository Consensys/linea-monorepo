package net.consensys.zkevm.encoding

import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1

fun interface ExecutionPayloadV1Encoder {
  fun encode(payload: ExecutionPayloadV1): ByteArray
}
