package linea.ftx

import linea.ftx.conflation.FtxConflationInfo
import java.util.Queue

class FtxProcessedQueuerForConflation(
  val conflationFtxQueue: Queue<FtxConflationInfo>,
  val aggregationFtxQueue: Queue<FtxConflationInfo>,
) : ForcedTransactionProcessedListener {

  /**
   * Add the ftx for conflation and aggregation queues.
   * It should add it for both queues or none of them
   */
  override fun onFtxProcessed(ftxStatus: FtxConflationInfo) {
    // Add the processed FTX to both queues (conflation and aggregation)
    if (conflationFtxQueue.offer(ftxStatus)) {}
    conflationFtxQueue.add(ftxStatus)
    aggregationFtxQueue.add(ftxStatus)
  }
}
