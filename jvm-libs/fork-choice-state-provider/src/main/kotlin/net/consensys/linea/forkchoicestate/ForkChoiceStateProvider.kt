package net.consensys.linea.forkchoicestate

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import java.util.concurrent.locks.ReadWriteLock
import java.util.concurrent.locks.ReentrantReadWriteLock

interface ForkChoiceStateProvider {
  fun getLatestForkChoiceState(): ForkChoiceStateInfoV0
}

class ForkChoiceStateController(initialForkChoiceState: ForkChoiceStateInfoV0) :
  ForkChoiceStateProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private var lastForkChoiceState: ForkChoiceStateInfoV0 = initialForkChoiceState
  private val readWriteLock: ReadWriteLock = ReentrantReadWriteLock()

  fun updateForkChoiceState(newState: ForkChoiceStateInfoV0) {
    try {
      readWriteLock.writeLock().lock()
      if (newState.headBlockNumber > lastForkChoiceState.headBlockNumber) {
        lastForkChoiceState = newState
      } else {
        log.error(
          "Trying to set head block to older block. Current head:{number={}, hash={}}, new head:{number={}, hash={}}",
          lastForkChoiceState.headBlockNumber,
          lastForkChoiceState.forkChoiceState.headBlockHash.toShortHexString(),
          newState.headBlockNumber,
          newState.forkChoiceState.headBlockHash.toShortHexString()
        )
      }
    } finally {
      readWriteLock.writeLock().unlock()
    }
  }

  fun updateHeadBlock(blockNumber: Long, blockTimestamp: Long, blockHash: Bytes32) {
    try {
      readWriteLock.writeLock().lock()
      val newState =
        lastForkChoiceState.copy(
          headBlockNumber = blockNumber,
          headBlockTimestamp = blockTimestamp,
          forkChoiceState = lastForkChoiceState.forkChoiceState.copy(headBlockHash = blockHash)
        )
      updateForkChoiceState(newState)
    } finally {
      readWriteLock.writeLock().unlock()
    }
  }

  override fun getLatestForkChoiceState(): ForkChoiceStateInfoV0 {
    val state =
      try {
        readWriteLock.readLock().lock()
        lastForkChoiceState
      } finally {
        readWriteLock.readLock().unlock()
      }
    return state
  }
}
