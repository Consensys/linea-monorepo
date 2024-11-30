package linea.staterecover

import build.linea.staterecover.BlockL1RecoveredData
import build.linea.staterecover.TransactionL1RecoveredData
import kotlinx.datetime.Instant
import net.consensys.decodeHex
import net.consensys.linea.blob.BlobDecompressor
import net.consensys.toULong
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.encoding.registry.BlockDecoder
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.RLP
import kotlin.jvm.optionals.getOrNull

interface BlobDecompressorAndDeserializer {
  /**
   * Decompresses the EIP4844 blobs and deserializes them into domain objects.
   */
  fun decompress(
    startBlockNumber: ULong,
    blobs: List<ByteArray>
  ): List<BlockL1RecoveredData>
}

data class BlockHeaderStaticFields(
  val coinbase: ByteArray,
  val gasLimit: ULong = 61_000_000UL,
  val difficulty: ULong = 2UL
) {
  companion object {
    val mainnet = BlockHeaderStaticFields(
      coinbase = "0x8F81e2E3F8b46467523463835F965fFE476E1c9E".decodeHex()
    )
    val sepolia = BlockHeaderStaticFields(
      coinbase = "0x4D517Aef039A48b3B6bF921e210b7551C8E37107".decodeHex()
    )
    val localDev = BlockHeaderStaticFields(
      coinbase = "0x6d976c9b8ceee705d4fe8699b44e5eb58242f484".decodeHex()
    )
  }
}

class BlobDecompressorToDomainV1(
  val decompressor: BlobDecompressor,
//  val chainId: ULong,
  val staticFields: BlockHeaderStaticFields
) : BlobDecompressorAndDeserializer {

  private val blockDecoder =
    BlockDecoder.builder().withTransactionDecoder({ NoSignatureTransactionDecoder() })
      .build()
  private val blockHeaderFunctions = MainnetBlockHeaderFunctions()

  override fun decompress(
    startBlockNumber: ULong,
    blobs: List<ByteArray>
  ): List<BlockL1RecoveredData> {
    val blocksRecovered = mutableListOf<BlockL1RecoveredData>()
    var blockNumber = startBlockNumber

    blobs.forEach { blob ->
      val blocksRlp = rlpDecodeAsListOfBytes(decompressor.decompress(blob))
      blocksRlp.forEach { blockRlp ->
        val block: Block = blockDecoder.decode(RLP.input(Bytes.wrap(blockRlp), true), blockHeaderFunctions)

        val blockRecovered = BlockL1RecoveredData(
          blockNumber = blockNumber++,
          blockHash = block.header.parentHash.toArray(),
          coinbase = staticFields.coinbase,
          blockTimestamp = Instant.fromEpochSeconds(block.header.timestamp),
          gasLimit = this.staticFields.gasLimit,
          difficulty = block.header.difficulty.asBigInteger.toULong(),
          transactions = block.body.transactions.map { transaction ->
            TransactionL1RecoveredData(
              type = transaction.type.serializedType.toUByte(),
              from = transaction.sender.toArray(),
              nonce = transaction.nonce.toULong(),
              gasLimit = transaction.gasLimit.toULong(),
              maxFeePerGas = transaction.maxFeePerGas.getOrNull()?.asBigInteger,
              maxPriorityFeePerGas = transaction.maxPriorityFeePerGas.getOrNull()?.asBigInteger,
              gasPrice = transaction.gasPrice.getOrNull()?.asBigInteger,
              to = transaction.to.getOrNull()?.toArray(),
              value = transaction.value.asBigInteger,
              data = transaction.data.getOrNull()?.toArray(),
              accessList = transaction.accessList.getOrNull()?.map { accessTuple ->
                TransactionL1RecoveredData.AccessTuple(
                  address = accessTuple.address.toArray(),
                  storageKeys = accessTuple.storageKeys.map { it.toArray() }
                )
              }
            )
          }
        )
        blocksRecovered.add(blockRecovered)
      }
    }

    return blocksRecovered
  }
}

internal fun rlpDecodeAsListOfBytes(rlpEncoded: ByteArray): List<ByteArray> {
  val decodedBytes = mutableListOf<ByteArray>()
  RLP.input(Bytes.wrap(rlpEncoded), true).also { rlpInput ->
    rlpInput.enterList()
    while (!rlpInput.isEndOfCurrentList) {
      decodedBytes.add(rlpInput.readBytes().toArray())
    }
    rlpInput.leaveList()
  }
  return decodedBytes
}
