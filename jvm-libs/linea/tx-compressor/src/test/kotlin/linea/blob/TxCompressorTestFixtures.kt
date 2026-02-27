package linea.blob

import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.crypto.KeyPair
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.datatypes.AccessListEntry
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.TransactionType
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput
import java.math.BigInteger
import kotlin.random.Random

/**
 * Holds the data needed to append a transaction to TxCompressor.
 * Format expected by the compressor: from (20 bytes) || rlpForSigning
 */
data class TransactionData(
  val from: ByteArray,
  val rlpForSigning: ByteArray,
) {
  val totalSize: Int get() = from.size + rlpForSigning.size

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false
    other as TransactionData
    return from.contentEquals(other.from) && rlpForSigning.contentEquals(other.rlpForSigning)
  }

  override fun hashCode(): Int {
    var result = from.contentHashCode()
    result = 31 * result + rlpForSigning.contentHashCode()
    return result
  }
}

/**
 * Shared transaction encoding and generation utilities for TxCompressor tests.
 *
 * All generators return [Pair<Transaction, TransactionData>] so callers can access both
 * the full Besu [Transaction] (for block building) and the pre-encoded [TransactionData]
 * (for the compressor).
 */
object TxCompressorTestFixtures {

  val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
  private val ERC20_TRANSFER_SELECTOR = Bytes.fromHexString("0xa9059cbb")

  // chainId used for EIP-1559 benchmark generators (Linea mainnet)
  private val LINEA_CHAIN_ID = BigInteger.valueOf(59144)

  // ── Encoding ──────────────────────────────────────────────────────────────

  /**
   * Encodes a transaction into the format expected by TxCompressor:
   * from (20 bytes) || RLP-for-signing (without signature).
   */
  fun encodeTransactionForCompressor(tx: Transaction): TransactionData {
    val from = tx.sender.toArray()
    val rlpForSigning = encodeTransactionForSigning(tx)
    return TransactionData(from, rlpForSigning)
  }

  /**
   * Encodes a transaction for signing (without the signature), matching the Go
   * ethereum.EncodeTxForSigning used by the v1 blob compressor.
   *
   * - FRONTIER EIP-155:  RLP([nonce, gasPrice, gasLimit, to, value, data, chainId, 0, 0])
   * - FRONTIER legacy:   RLP([nonce, gasPrice, gasLimit, to, value, data])
   * - ACCESS_LIST:       0x01 || RLP([chainId, nonce, gasPrice, gasLimit, to, value, data, accessList])
   * - EIP1559:           0x02 || RLP([chainId, nonce, maxPriorityFee, maxFee, gasLimit, to, value, data, accessList])
   */
  fun encodeTransactionForSigning(tx: Transaction): ByteArray {
    return when (tx.type) {
      TransactionType.FRONTIER -> encodeFrontierForSigning(tx)
      TransactionType.ACCESS_LIST -> byteArrayOf(0x01) + encodeAccessListForSigning(tx)
      TransactionType.EIP1559 -> byteArrayOf(0x02) + encodeEip1559ForSigning(tx)
      else -> throw IllegalArgumentException("Unsupported transaction type for signing: ${tx.type}")
    }
  }

  /** Returns the byte length of the standard signed wire-format RLP for [tx] (v/r/s included). */
  fun signedTxRlpSize(tx: Transaction): Int {
    val out = BytesValueRLPOutput()
    tx.writeTo(out)
    return out.encoded().size()
  }

  private fun encodeFrontierForSigning(tx: Transaction): ByteArray {
    val out = BytesValueRLPOutput()
    out.startList()
    out.writeLongScalar(tx.nonce)
    out.writeUInt256Scalar(tx.gasPrice.orElse(Wei.ZERO))
    out.writeLongScalar(tx.gasLimit)
    writeToAddress(out, tx)
    out.writeUInt256Scalar(tx.value)
    out.writeBytes(tx.payload)
    if (tx.chainId.isPresent) {
      out.writeBigIntegerScalar(tx.chainId.get())
      out.writeIntScalar(0)
      out.writeIntScalar(0)
    }
    out.endList()
    return out.encoded().toArray()
  }

  private fun encodeAccessListForSigning(tx: Transaction): ByteArray {
    val out = BytesValueRLPOutput()
    out.startList()
    out.writeBigIntegerScalar(tx.chainId.get())
    out.writeLongScalar(tx.nonce)
    out.writeUInt256Scalar(tx.gasPrice.orElse(Wei.ZERO))
    out.writeLongScalar(tx.gasLimit)
    writeToAddress(out, tx)
    out.writeUInt256Scalar(tx.value)
    out.writeBytes(tx.payload)
    writeAccessList(out, tx.accessList.orElse(emptyList()))
    out.endList()
    return out.encoded().toArray()
  }

  private fun encodeEip1559ForSigning(tx: Transaction): ByteArray {
    val out = BytesValueRLPOutput()
    out.startList()
    out.writeBigIntegerScalar(tx.chainId.get())
    out.writeLongScalar(tx.nonce)
    out.writeUInt256Scalar(tx.maxPriorityFeePerGas.orElse(Wei.ZERO))
    out.writeUInt256Scalar(tx.maxFeePerGas.orElse(Wei.ZERO))
    out.writeLongScalar(tx.gasLimit)
    writeToAddress(out, tx)
    out.writeUInt256Scalar(tx.value)
    out.writeBytes(tx.payload)
    writeAccessList(out, tx.accessList.orElse(emptyList()))
    out.endList()
    return out.encoded().toArray()
  }

  private fun writeToAddress(out: BytesValueRLPOutput, tx: Transaction) {
    if (tx.to.isPresent) {
      out.writeBytes(tx.to.get())
    } else {
      out.writeNull()
    }
  }

  private fun writeAccessList(out: BytesValueRLPOutput, accessList: List<AccessListEntry>) {
    out.startList()
    for (entry in accessList) {
      out.startList()
      out.writeBytes(entry.address)
      out.startList()
      for (key in entry.storageKeys()) {
        out.writeBytes(key)
      }
      out.endList()
      out.endList()
    }
    out.endList()
  }

  // ── FRONTIER (legacy) generators ─────────────────────────────────────────

  /**
   * Generates a randomized plain ETH transfer as a FRONTIER (EIP-155) transaction.
   * Each call uses a fresh key pair, so every transaction has a different sender.
   */
  fun generateRandomizedTransferTx(nonce: Long, random: Random): Pair<Transaction, TransactionData> {
    val keyPair = signatureAlgorithm.generateKeyPair()
    val toAddressBytes = ByteArray(20)
    random.nextBytes(toAddressBytes)
    val toAddress = Address.wrap(Bytes.wrap(toAddressBytes))
    val gasPrice = Wei.of(random.nextLong(1_000_000_000L, 100_000_000_000L))
    val gasLimit = random.nextLong(21000, 100000)
    val value = Wei.of(random.nextLong(0, 1_000_000_000_000_000_000L))
    val payloadSize = random.nextInt(0, 100)
    val payload = if (payloadSize > 0) {
      val payloadBytes = ByteArray(payloadSize)
      random.nextBytes(payloadBytes)
      Bytes.wrap(payloadBytes)
    } else {
      Bytes.EMPTY
    }
    val tx = Transaction.builder()
      .type(TransactionType.FRONTIER)
      .nonce(nonce)
      .gasPrice(gasPrice)
      .gasLimit(gasLimit)
      .to(toAddress)
      .value(value)
      .payload(payload)
      .chainId(BigInteger.ONE)
      .signAndBuild(keyPair)
    return tx to encodeTransactionForCompressor(tx)
  }

  /** Generates [count] randomized FRONTIER transfer transactions using a seeded [Random]. */
  fun generateManyRandomizedTransactions(count: Int, seed: Long = 12345L): List<Pair<Transaction, TransactionData>> {
    val random = Random(seed)
    return (0 until count).map { generateRandomizedTransferTx(it.toLong(), random) }
  }

  /**
   * Generates a single FRONTIER ERC-20 transfer transaction.
   * Caller controls the sender [keyPair], [tokenContractAddress], and [recipientAddress]
   * so tests can mix fixed and random values as needed.
   */
  fun generateErc20TransferTx(
    nonce: Long,
    keyPair: KeyPair,
    tokenContractAddress: Address,
    recipientAddress: Address,
    random: Random,
  ): Pair<Transaction, TransactionData> {
    val tokenAmount = BigInteger.valueOf(random.nextLong(1, Long.MAX_VALUE))
    val recipientPadded = Bytes.concatenate(
      Bytes.wrap(ByteArray(12)),
      recipientAddress,
    )
    val amountBytes = Bytes.wrap(tokenAmount.toByteArray())
    val amountPadded = Bytes.concatenate(
      Bytes.wrap(ByteArray(32 - amountBytes.size())),
      amountBytes,
    )
    val payload = Bytes.concatenate(ERC20_TRANSFER_SELECTOR, recipientPadded, amountPadded)
    val gasPrice = Wei.of(random.nextLong(1_000_000_000L, 50_000_000_000L))
    val tx = Transaction.builder()
      .type(TransactionType.FRONTIER)
      .nonce(nonce)
      .gasPrice(gasPrice)
      .gasLimit(100000)
      .to(tokenContractAddress)
      .value(Wei.ZERO)
      .payload(payload)
      .chainId(BigInteger.ONE)
      .signAndBuild(keyPair)
    return tx to encodeTransactionForCompressor(tx)
  }

  /** Generates [count] ERC-20 transfers from the same sender to the same recipient. */
  fun generateManyErc20Transfers(count: Int, seed: Long = 54321L): List<Pair<Transaction, TransactionData>> {
    val random = Random(seed)
    val keyPair = signatureAlgorithm.generateKeyPair()
    val tokenContractAddress = Address.fromHexString("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
    val recipientAddress = Address.fromHexString("0x1234567890123456789012345678901234567890")
    return (0 until count).map { i ->
      generateErc20TransferTx(i.toLong(), keyPair, tokenContractAddress, recipientAddress, random)
    }
  }

  // ── EIP-1559 generators (benchmark scenarios) ────────────────────────────

  /**
   * Generates a randomized plain ETH transfer as an EIP-1559 transaction.
   * Uses [LINEA_CHAIN_ID] (59144).
   */
  fun generateRandomizedEip1559PlainTransferTx(nonce: Long, random: Random): Pair<Transaction, TransactionData> {
    val keyPair = signatureAlgorithm.generateKeyPair()
    val toAddressBytes = ByteArray(20)
    random.nextBytes(toAddressBytes)
    val toAddress = Address.wrap(Bytes.wrap(toAddressBytes))
    val maxFeePerGas = Wei.of(random.nextLong(1_000_000_000L, 100_000_000_000L))
    val maxPriorityFeePerGas = Wei.of(random.nextLong(100_000_000L, 10_000_000_000L))
    val value = Wei.of(random.nextLong(0, 1_000_000_000_000_000_000L))
    val tx = Transaction.builder()
      .type(TransactionType.EIP1559)
      .nonce(nonce)
      .maxFeePerGas(maxFeePerGas)
      .maxPriorityFeePerGas(maxPriorityFeePerGas)
      .gasLimit(21000)
      .to(toAddress)
      .value(value)
      .payload(Bytes.EMPTY)
      .chainId(LINEA_CHAIN_ID)
      .signAndBuild(keyPair)
    return tx to encodeTransactionForCompressor(tx)
  }

  /**
   * Generates a randomized ERC-20 transfer as an EIP-1559 transaction.
   * Each call uses a fresh key pair and random contract/recipient addresses.
   * Uses [LINEA_CHAIN_ID] (59144).
   */
  fun generateRandomizedEip1559Erc20Tx(nonce: Long, random: Random): Pair<Transaction, TransactionData> {
    val keyPair = signatureAlgorithm.generateKeyPair()
    val contractBytes = ByteArray(20)
    random.nextBytes(contractBytes)
    val contractAddress = Address.wrap(Bytes.wrap(contractBytes))
    val recipientBytes = ByteArray(20)
    random.nextBytes(recipientBytes)
    val recipientAddress = Address.wrap(Bytes.wrap(recipientBytes))
    val tokenAmount = BigInteger.valueOf(random.nextLong(1, Long.MAX_VALUE))
    val recipientPadded = Bytes.concatenate(Bytes.wrap(ByteArray(12)), recipientAddress)
    val amountBytes = Bytes.wrap(tokenAmount.toByteArray())
    val amountPadded = Bytes.concatenate(Bytes.wrap(ByteArray(32 - amountBytes.size())), amountBytes)
    val payload = Bytes.concatenate(ERC20_TRANSFER_SELECTOR, recipientPadded, amountPadded)
    val maxFeePerGas = Wei.of(random.nextLong(1_000_000_000L, 100_000_000_000L))
    val maxPriorityFeePerGas = Wei.of(random.nextLong(100_000_000L, 10_000_000_000L))
    val tx = Transaction.builder()
      .type(TransactionType.EIP1559)
      .nonce(nonce)
      .maxFeePerGas(maxFeePerGas)
      .maxPriorityFeePerGas(maxPriorityFeePerGas)
      .gasLimit(100000)
      .to(contractAddress)
      .value(Wei.ZERO)
      .payload(payload)
      .chainId(LINEA_CHAIN_ID)
      .signAndBuild(keyPair)
    return tx to encodeTransactionForCompressor(tx)
  }

  /**
   * Generates a randomized contract call as an EIP-1559 transaction with [calldataSize] bytes.
   * Uses [LINEA_CHAIN_ID] (59144).
   */
  fun generateRandomizedEip1559CalldataTx(nonce: Long, calldataSize: Int, random: Random): Pair<Transaction, TransactionData> {
    val keyPair = signatureAlgorithm.generateKeyPair()
    val contractBytes = ByteArray(20)
    random.nextBytes(contractBytes)
    val contractAddress = Address.wrap(Bytes.wrap(contractBytes))
    val calldataBytes = ByteArray(calldataSize)
    random.nextBytes(calldataBytes)
    val maxFeePerGas = Wei.of(random.nextLong(1_000_000_000L, 100_000_000_000L))
    val maxPriorityFeePerGas = Wei.of(random.nextLong(100_000_000L, 10_000_000_000L))
    val gasLimit = minOf(100_000L + calldataSize.toLong() * 16L, 30_000_000L)
    val tx = Transaction.builder()
      .type(TransactionType.EIP1559)
      .nonce(nonce)
      .maxFeePerGas(maxFeePerGas)
      .maxPriorityFeePerGas(maxPriorityFeePerGas)
      .gasLimit(gasLimit)
      .to(contractAddress)
      .value(Wei.ZERO)
      .payload(Bytes.wrap(calldataBytes))
      .chainId(LINEA_CHAIN_ID)
      .signAndBuild(keyPair)
    return tx to encodeTransactionForCompressor(tx)
  }
}
