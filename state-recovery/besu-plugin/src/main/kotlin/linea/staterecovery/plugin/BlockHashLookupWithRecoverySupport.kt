package linea.staterecovery.plugin

import net.consensys.encodeHex
import net.consensys.minusCoercingUnderflow
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.evm.blockhash.BlockHashLookup
import org.hyperledger.besu.evm.frame.MessageFrame
import java.util.concurrent.ConcurrentHashMap

class BlockHashLookupWithRecoverySupport(
  val lookbackWindow: ULong,
  private val log: Logger = LogManager.getLogger(BlockHashLookupWithRecoverySupport::class.java)
) : BlockHashLookup {
  private val lookbackHashesMap = ConcurrentHashMap<ULong, ByteArray>()

  fun areSequential(numbers: List<ULong>): Boolean {
    if (numbers.size < 2) return true // A list with less than 2 elements is trivially continuous

    for (i in 1 until numbers.size) {
      if (numbers[i] != numbers[i - 1] + 1UL) {
        return false
      }
    }
    return true
  }

  fun addLookbackHashes(blocksHashes: Map<ULong, ByteArray>) {
    require(areSequential(blocksHashes.keys.toList())) {
      "Block numbers must be sequential"
    }

    log.debug("adding hashes: {}", {
      blocksHashes.toList()
        .sortedBy { it.first }
        .map { (k, v) -> "$k=${v.encodeHex()}" }
    })

    lookbackHashesMap.putAll(blocksHashes)
  }

  fun addHeadBlockHash(blockNumber: ULong, blockHash: ByteArray) {
    lookbackHashesMap[blockNumber] = blockHash
    pruneLookBackHashes(blockNumber)
  }

  fun pruneLookBackHashes(headBlockNumber: ULong) {
    if (headBlockNumber <= lookbackWindow) return

    lookbackHashesMap.keys.removeIf {
      // <= would wrongly remove block 0 while headBlockNumber < lookBackWindow,
      // we keep 1 block more that we need, but it's fine
      it < headBlockNumber.minusCoercingUnderflow(lookbackWindow)
    }
  }

  fun getHash(blockNumber: Long): Hash {
    val resolvedHash = lookbackHashesMap[blockNumber.toULong()]
      ?.let { Hash.wrap(Bytes32.wrap(it)) }
      ?: Hash.ZERO

    log.debug("block={} lookback hash={}", blockNumber, resolvedHash)
    return resolvedHash
  }

  override fun apply(t: MessageFrame?, blockNumber: Long): Hash {
    return getHash(blockNumber)
  }
}
