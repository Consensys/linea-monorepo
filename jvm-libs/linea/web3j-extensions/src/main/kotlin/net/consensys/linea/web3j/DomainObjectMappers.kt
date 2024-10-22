package net.consensys.linea.web3j

import net.consensys.linea.bigIntFromPrefixedHex
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
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
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger

private val log: Logger = LogManager.getLogger("DomainObjectMappers")
fun EthBlock.Block.toExecutionPayloadV1(): ExecutionPayloadV1 {
  /**
   * @JsonProperty("parentHash") Bytes32 parentHash,
   * @JsonProperty("feeRecipient") Bytes20 feeRecipient,
   * @JsonProperty("stateRoot") Bytes32 stateRoot,
   * @JsonProperty("receiptsRoot") Bytes32 receiptsRoot,
   * @JsonProperty("logsBloom") Bytes logsBloom,
   * @JsonProperty("prevRandao") Bytes32 prevRandao,
   * @JsonProperty("blockNumber") UInt64 blockNumber,
   * @JsonProperty("gasLimit") UInt64 gasLimit,
   * @JsonProperty("gasUsed") UInt64 gasUsed,
   * @JsonProperty("timestamp") UInt64 timestamp,
   * @JsonProperty("extraData") Bytes extraData,
   * @JsonProperty("baseFeePerGas") UInt256 baseFeePerGas,
   * @JsonProperty("blockHash") Bytes32 blockHash,
   * @JsonProperty("transactions") List<Bytes> transactions)
   */
  return ExecutionPayloadV1(
    Bytes32.fromHexString(this.parentHash),
    Bytes20.fromHexString(this.miner),
    Bytes32.fromHexString(this.stateRoot),
    Bytes32.fromHexString(this.receiptsRoot),
    Bytes.fromHexString(this.logsBloom),
    Bytes32.fromHexString(this.mixHash),
    UInt64.valueOf(this.number),
    UInt64.valueOf(this.gasLimit),
    UInt64.valueOf(this.gasUsed),
    UInt64.valueOf(this.timestamp),
    Bytes.fromHexString(this.extraData),
    UInt256.valueOf(this.baseFeePerGas),
    Bytes32.fromHexString(this.hash),
    this.transactions.map {
      val transaction = it.get() as EthBlock.TransactionObject
      kotlin.runCatching {
        transaction.toBytes()
      }.onFailure { th ->
        log.error(
          "Failed to encode transaction! blockNumber={} tx={} errorMessage={}",
          this.number,
          transaction.hash.toString(),
          th.message,
          th
        )
      }
        .getOrThrow()
    }
  )
}

fun recIdFromV(v: BigInteger): Pair<Byte, BigInteger?> {
  val recId: Byte
  var chainId: BigInteger? = null
  if (v == Transaction.REPLAY_UNPROTECTED_V_BASE || v == Transaction.REPLAY_UNPROTECTED_V_BASE_PLUS_1) {
    recId = v.subtract(Transaction.REPLAY_UNPROTECTED_V_BASE).byteValueExact()
  } else if (v > Transaction.REPLAY_PROTECTED_V_MIN) {
    chainId = v.subtract(Transaction.REPLAY_PROTECTED_V_BASE).divide(Transaction.TWO)
    recId = v.subtract(Transaction.TWO.multiply(chainId).add(Transaction.REPLAY_PROTECTED_V_BASE)).byteValueExact()
  } else {
    throw RuntimeException("An unsupported encoded `v` value of $v was found")
  }
  return Pair(recId, chainId)
}

// TODO: Test
fun EthBlock.TransactionObject.toBytes(): Bytes {
  val isFrontier = this.type == "0x0"
  val (recId, chainId) = if (isFrontier) {
    recIdFromV(this.v.toBigInteger())
  } else {
    Pair(this.v.toByte(), BigInteger.valueOf(this.chainId))
  }
  val signature = SECPSignature.create(
    this.r.bigIntFromPrefixedHex(),
    this.s.bigIntFromPrefixedHex(),
    recId,
    SecP256K1Curve().order
  )

  val transaction = Transaction.builder()
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
              entry.storageKeys
            )
          }
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
