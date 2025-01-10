package linea.staterecover

interface RecoveryStatusPersistence {
  fun saveRecoveryStartBlockNumber(recoveryStartBlockNumber: ULong)
  fun getRecoveryStartBlockNumber(): ULong?
}

class InMemoryRecoveryStatus : RecoveryStatusPersistence {
  private var recoveryStartBlockNumber: ULong? = null

  @Synchronized
  override fun saveRecoveryStartBlockNumber(recoveryStartBlockNumber: ULong) {
    this.recoveryStartBlockNumber = recoveryStartBlockNumber
  }

  @Synchronized
  override fun getRecoveryStartBlockNumber(): ULong? {
    return recoveryStartBlockNumber
  }
}
