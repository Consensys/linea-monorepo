/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.executionlayer.extensions

import maru.core.ExecutionPayload
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64

fun ExecutionPayloadV3.toDomainExecutionPayload() =
  ExecutionPayload(
    parentHash = this.parentHash.toArray(),
    feeRecipient = this.feeRecipient.wrappedBytes.toArray(),
    stateRoot = this.stateRoot.toArray(),
    receiptsRoot = this.receiptsRoot.toArray(),
    logsBloom = this.logsBloom.toArray(),
    prevRandao = this.prevRandao.toArray(),
    blockNumber = this.blockNumber.longValue().toULong(),
    gasLimit = this.gasLimit.longValue().toULong(),
    gasUsed = this.gasUsed.longValue().toULong(),
    timestamp = this.timestamp.longValue().toULong(),
    extraData = this.extraData.toArray(),
    baseFeePerGas =
      this.baseFeePerGas.toBigInteger(),
    blockHash = this.blockHash.toArray(),
    transactions = this.transactions.map { it.toArray() },
  )

fun ExecutionPayload.toExecutionPayloadV3() =
  ExecutionPayloadV3(
    /* parentHash */ Bytes32.wrap(this.parentHash),
    /* feeRecipient */ Bytes20(Bytes.wrap(this.feeRecipient)),
    /* stateRoot */ Bytes32.wrap(this.stateRoot),
    /* receiptsRoot */ Bytes32.wrap(this.receiptsRoot),
    /* logsBloom */ Bytes.wrap(this.logsBloom),
    /* prevRandao */ Bytes32.wrap(this.prevRandao),
    /* blockNumber */ UInt64.valueOf(this.blockNumber.toString()),
    /* gasLimit */ UInt64.valueOf(this.gasLimit.toString()),
    /* gasUsed */ UInt64.valueOf(this.gasUsed.toString()),
    /* timestamp */ UInt64.valueOf(this.timestamp.toString()),
    /* extraData */ Bytes.wrap(this.extraData),
    /* baseFeePerGas */ UInt256.valueOf(this.baseFeePerGas),
    /* blockHash */ Bytes32.wrap(this.blockHash),
    /* transactions */ this.transactions.map { Bytes.wrap(it) },
    /* withdrawals */ emptyList(),
    /* blobGasUsed */ UInt64.ZERO,
    /* excessBlobGas */ UInt64.ZERO,
  )
