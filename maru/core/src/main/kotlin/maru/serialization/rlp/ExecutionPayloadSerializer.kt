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
package maru.serialization.rlp

import java.math.BigInteger
import maru.core.ExecutionPayload
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class ExecutionPayloadSerializer : RLPSerializer<ExecutionPayload> {
  override fun writeTo(
    value: ExecutionPayload,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()

    rlpOutput.writeBytes(Bytes.wrap(value.parentHash))
    rlpOutput.writeBytes(Bytes.wrap(value.feeRecipient))
    rlpOutput.writeBytes(Bytes.wrap(value.stateRoot))
    rlpOutput.writeBytes(Bytes.wrap(value.receiptsRoot))
    rlpOutput.writeBytes(Bytes.wrap(value.logsBloom))
    rlpOutput.writeBytes(Bytes.wrap(value.prevRandao))
    rlpOutput.writeLong(value.blockNumber.toLong())
    rlpOutput.writeLong(value.gasLimit.toLong())
    rlpOutput.writeLong(value.gasUsed.toLong())
    rlpOutput.writeLong(value.timestamp.toLong())
    rlpOutput.writeBytes(Bytes.wrap(value.extraData))
    rlpOutput.writeBytes(Bytes.wrap(value.baseFeePerGas.toByteArray()))
    rlpOutput.writeBytes(Bytes.wrap(value.blockHash))
    rlpOutput.writeList(value.transactions) { transaction, output ->
      output.writeBytes(Bytes.wrap(transaction))
    }
    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): ExecutionPayload {
    rlpInput.enterList()
    val parentHash = rlpInput.readBytes().toArray()
    val feeRecipient = rlpInput.readBytes().toArray()
    val stateRoot = rlpInput.readBytes().toArray()
    val receiptsRoot = rlpInput.readBytes().toArray()
    val logsBloom = rlpInput.readBytes().toArray()
    val prevRandao = rlpInput.readBytes().toArray()
    val blockNumber = rlpInput.readLong().toULong()
    val gasLimit = rlpInput.readLong().toULong()
    val gasUsed = rlpInput.readLong().toULong()
    val timestamp = rlpInput.readLong().toULong()
    val extraData = rlpInput.readBytes().toArray()
    val baseFeePerGas = BigInteger(rlpInput.readBytes().toArray())
    val blockHash = rlpInput.readBytes().toArray()
    val transactions = rlpInput.readList { it.readBytes().toArray() }.toList()
    rlpInput.leaveList()
    return ExecutionPayload(
      parentHash = parentHash,
      feeRecipient = feeRecipient,
      stateRoot = stateRoot,
      receiptsRoot = receiptsRoot,
      logsBloom = logsBloom,
      prevRandao = prevRandao,
      blockNumber = blockNumber,
      gasLimit = gasLimit,
      gasUsed = gasUsed,
      timestamp = timestamp,
      extraData = extraData,
      baseFeePerGas = baseFeePerGas,
      blockHash = blockHash,
      transactions = transactions,
    )
  }
}
