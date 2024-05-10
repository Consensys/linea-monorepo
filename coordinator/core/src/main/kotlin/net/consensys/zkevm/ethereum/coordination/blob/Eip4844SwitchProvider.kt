package net.consensys.zkevm.ethereum.coordination.blob

import net.consensys.zkevm.domain.Blob

interface Eip4844SwitchProvider {
  fun isEip4844Enabled(blob: Blob): Boolean
}

class Eip4844SwitchProviderImpl(private val blockNumberToSwitch: Long) : Eip4844SwitchProvider {

  fun isEip4844Enabled(startBlockNumber: ULong): Boolean {
    return if (blockNumberToSwitch < 0) {
      false
    } else {
      startBlockNumber.toLong() >= blockNumberToSwitch
    }
  }
  override fun isEip4844Enabled(blob: Blob): Boolean {
    return isEip4844Enabled(blob.startBlockNumber)
  }
}
