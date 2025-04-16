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
package maru.e2e

import fromHexToByteArray
import java.math.BigInteger
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.ExecutionPayload
import maru.core.HashUtil
import maru.core.Validator
import maru.serialization.rlp.KeccakHasher
import maru.serialization.rlp.RLPSerializers
import org.apache.tuweni.bytes.Bytes
import org.bouncycastle.math.ec.custom.sec.SecP256K1Curve
import org.hyperledger.besu.crypto.SECPSignature
import org.hyperledger.besu.datatypes.AccessListEntry
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.core.encoding.EncodingContext
import org.hyperledger.besu.ethereum.core.encoding.TransactionEncoder
import org.web3j.protocol.core.methods.response.EthBlock
import org.web3j.protocol.core.methods.response.EthBlock.TransactionObject

// TODO: This is a copypaste from
// https://github.com/Consensys/linea-monorepo/blob/main/jvm-libs/linea/web3j-extensions/src/main/kotlin/net/consensys/linea/web3j/DomainObjectMappers.kt clean up later
object Mappers {
  private val hasher = HashUtil.headerHash(RLPSerializers.BeaconBlockHeaderSerializer, KeccakHasher)

  private fun recIdFromV(v: BigInteger): Pair<Byte, BigInteger?> {
    val recId: Byte
    var chainId: BigInteger? = null
    if (
      v == Transaction.REPLAY_UNPROTECTED_V_BASE ||
      v == Transaction.REPLAY_UNPROTECTED_V_BASE_PLUS_1
    ) {
      recId = v.subtract(Transaction.REPLAY_UNPROTECTED_V_BASE).byteValueExact()
    } else if (v > Transaction.REPLAY_PROTECTED_V_MIN) {
      chainId = v.subtract(Transaction.REPLAY_PROTECTED_V_BASE).divide(Transaction.TWO)
      recId =
        v
          .subtract(Transaction.TWO.multiply(chainId).add(Transaction.REPLAY_PROTECTED_V_BASE))
          .byteValueExact()
    } else {
      throw RuntimeException("An unsupported encoded `v` value of $v was found")
    }
    return Pair(recId, chainId)
  }

  // TODO: Test
  private fun TransactionObject.toByteArray(): ByteArray {
    val isFrontier = this.type == "0x0"
    val (recId, chainId) =
      if (isFrontier) {
        recIdFromV(this.v.toBigInteger())
      } else {
        Pair(this.v.toByte(), BigInteger.valueOf(this.chainId))
      }
    val signature =
      SECPSignature.create(
        BigInteger(this.r.removePrefix("0x"), 16),
        BigInteger(this.s.removePrefix("0x"), 16),
        recId,
        SecP256K1Curve().order,
      )

    val transaction =
      Transaction
        .builder()
        .nonce(this.nonce.toLong())
        .also { builder ->
          if (isFrontier || this.type == "0x1") {
            builder.gasPrice(Wei.of(this.gasPrice))
          } else {
            builder.maxPriorityFeePerGas(Wei.of(this.maxPriorityFeePerGas))
            builder.maxFeePerGas(Wei.of(this.maxFeePerGas))
          }
        }.gasLimit(this.gas.toLong())
        .to(Address.fromHexString(this.to))
        .value(Wei.of(this.value))
        .signature(signature)
        .payload(Bytes.fromHexString(this.input))
        .also { builder ->
          this.accessList?.also { accessList ->
            builder.accessList(
              accessList.map { entry ->
                AccessListEntry.createAccessListEntry(
                  Address.fromHexString(entry.address),
                  entry.storageKeys,
                )
              },
            )
          }
        }.sender(Address.fromHexString(this.from))
        .apply {
          if (chainId != null) {
            chainId(chainId)
          }
        }.build()

    return TransactionEncoder.encodeOpaqueBytes(transaction, EncodingContext.BLOCK_BODY).toArray()
  }

  fun EthBlock.Block.toDomain(): BeaconBlock {
    val transactions =
      this.transactions.map {
        val transaction = it.get() as TransactionObject
        transaction.toByteArray()
      }

    val executionPayload =
      ExecutionPayload(
        parentHash = this.parentHash.fromHexToByteArray(),
        feeRecipient = this.miner.fromHexToByteArray(),
        stateRoot = this.stateRoot.fromHexToByteArray(),
        receiptsRoot = this.receiptsRoot.fromHexToByteArray(),
        logsBloom = this.logsBloom.fromHexToByteArray(),
        prevRandao = this.mixHash.fromHexToByteArray(),
        blockNumber = this.number.toLong().toULong(),
        gasLimit = this.gasLimit.toLong().toULong(),
        gasUsed = this.gasUsed.toLong().toULong(),
        timestamp = this.timestamp.toLong().toULong(),
        extraData = this.extraData.fromHexToByteArray(),
        baseFeePerGas = this.baseFeePerGas,
        blockHash = this.hash.fromHexToByteArray(),
        transactions = transactions, // Transactions are omitted
      )
    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptyList(), executionPayload = executionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = this.timestamp.toLong().toULong(),
        proposer = Validator(this.miner.fromHexToByteArray()),
        parentRoot = ByteArray(32),
        stateRoot = ByteArray(32),
        bodyRoot = ByteArray(32),
        headerHashFunction = hasher,
      )
    return BeaconBlock(beaconBlockHeader, beaconBlockBody)
  }
}
