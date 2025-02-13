package linea.staterecovery.plugin

import linea.staterecovery.BlockFromL1RecoveredData
import net.consensys.encodeHex
import net.consensys.toBigInteger
import net.consensys.toULong
import org.apache.logging.log4j.LogManager
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.StateOverrideMap
import org.hyperledger.besu.evm.blockhash.BlockHashLookup
import org.hyperledger.besu.evm.frame.MessageFrame
import org.hyperledger.besu.plugin.data.BlockContext
import org.hyperledger.besu.plugin.data.BlockHeader
import org.hyperledger.besu.plugin.data.BlockOverrides
import org.hyperledger.besu.plugin.data.PluginBlockSimulationResult
import org.hyperledger.besu.plugin.services.BlockSimulationService
import org.hyperledger.besu.plugin.services.BlockchainService
import org.hyperledger.besu.plugin.services.sync.SynchronizationService
import java.util.concurrent.ConcurrentHashMap

class BlockHashLookupWithRecoverySupport(
  val lookbackWindow: ULong = 256UL
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

    lookbackHashesMap.putAll(blocksHashes)
  }

  fun addHeadBlockHash(blockNumber: ULong, blockHash: ByteArray) {
    lookbackHashesMap[blockNumber] = blockHash
    pruneLookBackHashes(blockNumber)
  }

  fun pruneLookBackHashes(headBlockNumber: ULong) {
    lookbackHashesMap.keys.removeIf { it < headBlockNumber - lookbackWindow }
  }

  fun getHash(blockNumber: Long): Hash {
    return lookbackHashesMap[blockNumber.toULong()]
      ?.let { Hash.wrap(Bytes32.wrap(it)) }
      ?: Hash.ZERO
  }

  override fun apply(t: MessageFrame?, blockNumber: Long): Hash {
    return getHash(blockNumber)
  }
}

class BlockImporter(
  private val blockchainService: BlockchainService,
  private val simulatorService: BlockSimulationService,
  private val synchronizationService: SynchronizationService,
  private val blockHashLookup: BlockHashLookupWithRecoverySupport = BlockHashLookupWithRecoverySupport()
) {
  private val log = LogManager.getLogger(BlockImporter::class.java)
  private val chainId = blockchainService.chainId.orElseThrow().toULong()

  fun addLookbackHashes(blocksHashes: Map<ULong, ByteArray>) {
    blockHashLookup.addLookbackHashes(blocksHashes)
  }

  fun importBlock(block: BlockFromL1RecoveredData): PluginBlockSimulationResult {
    val executedBlockResult = executeBlockWithTransactionsWithoutSignature(block)
    val result = importBlock(BlockContextData(executedBlockResult.blockHeader, executedBlockResult.blockBody))
    blockHashLookup.addHeadBlockHash(block.header.blockNumber, block.header.blockHash)
    return result
  }

  private fun executeBlockWithTransactionsWithoutSignature(
    block: BlockFromL1RecoveredData
  ): PluginBlockSimulationResult {
    log.trace(
      "simulating import block={} blockHash={}",
      block.header.blockNumber,
      block.header.blockHash.encodeHex()
    )
    val transactions = TransactionMapper.mapToBesu(
      block.transactions,
      chainId
    )
    val parentBlockNumber = block.header.blockNumber.toLong() - 1

    val executedBlockResult =
      simulatorService.simulate(
        parentBlockNumber,
        transactions,
        createOverrides(block, blockHashLookup::getHash),
        StateOverrideMap()
      )

    log.trace(
      " import simulation result: block={} blockHeader={}",
      executedBlockResult.blockHeader.number,
      executedBlockResult.blockHeader
    )
    return executedBlockResult
  }

  fun importBlock(context: BlockContext): PluginBlockSimulationResult {
    log.trace(
      "calling simulateAndPersistWorldState block={} blockHeader={}",
      context.blockHeader.number,
      context.blockHeader
    )
    val parentBlockNumber = context.blockHeader.number - 1
    val importedBlockResult =
      simulatorService.simulateAndPersistWorldState(
        parentBlockNumber,
        context.blockBody.transactions,
        createOverrides(context.blockHeader, blockHashLookup::getHash),
        StateOverrideMap()
      )
    log.trace(
      "simulateAndPersistWorldState result: block={} blockHeader={}",
      context.blockHeader.number,
      importedBlockResult.blockHeader
    )
    storeAndSetHead(importedBlockResult)
    return importedBlockResult
  }

  private fun storeAndSetHead(block: PluginBlockSimulationResult) {
    log.debug(
      "storeAndSetHead result: blockHeader={}",
      block.blockHeader
    )
    blockchainService.storeBlock(
      block.blockHeader,
      block.blockBody,
      block.receipts
    )
    synchronizationService.setHeadUnsafe(block.blockHeader, block.blockBody)
  }

  companion object {
    fun createOverrides(
      blockFromBlob: BlockFromL1RecoveredData,
      blockHashLookup: (Long) -> Hash
    ): BlockOverrides {
      return BlockOverrides.builder()
        .blockHash(Hash.wrap(Bytes32.wrap(blockFromBlob.header.blockHash)))
        .feeRecipient(Address.fromHexString(blockFromBlob.header.coinbase.encodeHex()))
        .blockNumber(blockFromBlob.header.blockNumber.toLong())
        .gasLimit(blockFromBlob.header.gasLimit.toLong())
        .timestamp(blockFromBlob.header.blockTimestamp.epochSeconds)
        .difficulty(blockFromBlob.header.difficulty.toBigInteger())
        .mixHashOrPrevRandao(Hash.ZERO)
        .blockHashLookup(blockHashLookup)
        .build()
    }

    fun createOverrides(
      blockHeader: BlockHeader,
      blockHashLookup: (Long) -> Hash
    ): BlockOverrides {
      return BlockOverrides.builder()
        .feeRecipient(blockHeader.coinbase)
        .blockNumber(blockHeader.number)
        .gasLimit(blockHeader.gasLimit)
        .timestamp(blockHeader.timestamp)
        .difficulty(blockHeader.difficulty.asBigInteger)
        .stateRoot(blockHeader.stateRoot)
        .mixHashOrPrevRandao(Hash.ZERO)
        .blockHashLookup(blockHashLookup)
        .build()
    }
  }
}
