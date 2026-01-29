package linea.ftx

import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import linea.forcedtx.ForcedTransactionRequest
import linea.forcedtx.ForcedTransactionResponse
import linea.forcedtx.ForcedTransactionsClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.CopyOnWriteArrayList
import kotlin.time.Clock

class FakeForcedTransactionsClient() : ForcedTransactionsClient {
  val ftxReceived = CopyOnWriteArrayList<ForcedTransactionRequest>()
  val ftxInclusionResults: MutableMap<ULong, ForcedTransactionInclusionStatus> = ConcurrentHashMap()
  val ftxInclusionResultsAfterReception: MutableMap<ULong, ForcedTransactionInclusionStatus> = ConcurrentHashMap()

  val ftxReceivedIds: List<ULong>
    get() = ftxReceived.map { it.ftxNumber }

  private fun fakeInclusionStatus(
    ftxNumber: ULong,
    l2BlockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult,
  ): ForcedTransactionInclusionStatus {
    return ForcedTransactionInclusionStatus(
      ftxNumber = ftxNumber,
      blockNumber = l2BlockNumber,
      blockTimestamp = Clock.System.now(),
      inclusionResult = inclusionResult,
      ftxHash = ByteArray(0),
      from = ByteArray(0),
    )
  }

  fun setFtxInclusionResult(
    ftxNumber: ULong,
    l2BlockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult,
  ) {
    ftxInclusionResults[ftxNumber] = fakeInclusionStatus(ftxNumber, l2BlockNumber, inclusionResult)
  }

  fun setFtxInclusionResultAfterReception(
    ftxNumber: ULong,
    l2BlockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult,
  ) {
    ftxInclusionResultsAfterReception[ftxNumber] = fakeInclusionStatus(ftxNumber, l2BlockNumber, inclusionResult)
  }

  override fun lineaSendForcedRawTransaction(
    transactions: List<ForcedTransactionRequest>,
  ): SafeFuture<List<ForcedTransactionResponse>> {
    ftxReceived.addAll(transactions)
    val results = transactions
      .map {
        ForcedTransactionResponse(
          ftxNumber = it.ftxNumber,
          ftxHash = it.ftxRlp.copyOfRange(0, it.ftxRlp.size.coerceAtMost(31)),
          ftxError = null,
        )
      }

    return SafeFuture.completedFuture(results)
      .thenPeek {
        transactions.forEach { ftx ->
          ftxInclusionResultsAfterReception[ftx.ftxNumber]
            ?.let { ftxInclusionResult ->
              ftxInclusionResults[ftx.ftxNumber] = ftxInclusionResult
              ftxInclusionResultsAfterReception.remove(ftx.ftxNumber)
            }
        }
      }
  }

  override fun lineaFindForcedTransactionStatus(ftxNumber: ULong): SafeFuture<ForcedTransactionInclusionStatus?> {
    return SafeFuture.completedFuture(ftxInclusionResults[ftxNumber])
  }
}
