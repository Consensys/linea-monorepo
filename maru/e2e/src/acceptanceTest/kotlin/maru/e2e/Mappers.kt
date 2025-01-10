package maru.e2e

import java.math.BigInteger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
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
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.WithdrawalV1
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64

// TODO: This is a copypaste from
// https://github.com/Consensys/linea-monorepo/blob/main/jvm-libs/linea/web3j-extensions/src/main/kotlin/net/consensys/linea/web3j/DomainObjectMappers.kt clean up later
object Mappers {
  private val ExtraDataStringHexLength = 2 + 32 * 2 // Prefix + 2 hexchars per 1 byte

  fun recIdFromV(v: BigInteger): Pair<Byte, BigInteger?> {
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
        v.subtract(Transaction.TWO.multiply(chainId).add(Transaction.REPLAY_PROTECTED_V_BASE))
          .byteValueExact()
    } else {
      throw RuntimeException("An unsupported encoded `v` value of $v was found")
    }
    return Pair(recId, chainId)
  }

  // TODO: Test
  private fun TransactionObject.toBytes(): Bytes {
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
      Transaction.builder()
        .nonce(this.nonce.toLong())
        .also { builder ->
          if (isFrontier || this.type == "0x1") {
            builder.gasPrice(Wei.of(this.gasPrice))
          } else {
            builder.maxPriorityFeePerGas(Wei.of(this.maxPriorityFeePerGas))
            builder.maxFeePerGas(Wei.of(this.maxFeePerGas))
          }
        }
        .gasLimit(this.gas.toLong())
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
        }
        .sender(Address.fromHexString(this.from))
        .apply {
          if (chainId != null) {
            chainId(chainId)
          }
        }
        .build()

    return TransactionEncoder.encodeOpaqueBytes(transaction, EncodingContext.BLOCK_BODY)
  }

  fun executionPayloadV3FromBlock(block: EthBlock.Block): ExecutionPayloadV3 {
    val parentHash = Bytes32.fromHexString(block.parentHash)
    val feeRecipient = Bytes20.fromHexString(block.miner)
    val stateRoot = Bytes32.fromHexString(block.stateRoot)
    val receiptsRoot = Bytes32.fromHexString(block.receiptsRoot)
    val logsBloom = Bytes.fromHexString(block.logsBloom)
    val prevRandao = Bytes32.fromHexString(block.mixHash)
    val blockNumber = UInt64.valueOf(block.number)
    val gasLimit = UInt64.valueOf(block.gasLimit)
    val gasUsed = UInt64.valueOf(block.gasUsed)
    val timestamp = UInt64.valueOf(block.timestamp)
    val extraData = Bytes.fromHexString(block.extraData)
    val baseFeePerGas = UInt256.valueOf(block.baseFeePerGas)
    val blockHash = Bytes32.fromHexString(block.hash)
    val transactions =
      block.transactions.map {
        val transaction = it.get() as TransactionObject
        transaction.toBytes()
      }
    val withdrawals = emptyList<WithdrawalV1>()
    val blobGasUsed = UInt64.valueOf(block.blobGasUsed)
    val excessBlobGas = UInt64.valueOf(block.excessBlobGas)

    // Create an instance of ExecutionPayloadV3
    return ExecutionPayloadV3(
      parentHash,
      feeRecipient,
      stateRoot,
      receiptsRoot,
      logsBloom,
      prevRandao,
      blockNumber,
      gasLimit,
      gasUsed,
      timestamp,
      extraData,
      baseFeePerGas,
      blockHash,
      transactions,
      withdrawals,
      blobGasUsed,
      excessBlobGas,
    )
  }

  fun executionPayloadV1FromBlock(block: EthBlock.Block): ExecutionPayloadV1 {
    val parentHash = Bytes32.fromHexString(block.parentHash)
    val feeRecipient = Bytes20.fromHexString(block.miner)
    val stateRoot = Bytes32.fromHexString(block.stateRoot)
    val receiptsRoot = Bytes32.fromHexString(block.receiptsRoot)
    val logsBloom = Bytes.fromHexString(block.logsBloom)
    val prevRandao = Bytes32.fromHexString(block.mixHash)
    val blockNumber = UInt64.valueOf(block.number)
    val gasLimit = UInt64.valueOf(block.gasLimit)
    val gasUsed = UInt64.valueOf(block.gasUsed)
    val timestamp = UInt64.valueOf(block.timestamp)
    //        val extraData = Bytes.fromHexString(block.extraData.substring(0,
    // ExtraDataStringHexLength))
    val extraData = Bytes.fromHexString(block.extraData)
    val baseFeePerGas = UInt256.valueOf(block.baseFeePerGas)
    val blockHash = Bytes32.fromHexString(block.hash)

    val transactions =
      block.transactions.map {
        val transaction = it.get() as TransactionObject
        transaction.toBytes()
      }

    // Create an instance of ExecutionPayloadV3
    return ExecutionPayloadV1(
      parentHash,
      feeRecipient,
      stateRoot,
      receiptsRoot,
      logsBloom,
      prevRandao,
      blockNumber,
      gasLimit,
      gasUsed,
      timestamp,
      extraData,
      baseFeePerGas,
      blockHash,
      transactions,
    )
  }
}
