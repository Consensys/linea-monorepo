package linea.ftx

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import linea.contract.events.FinalizedStateUpdatedEvent
import linea.contract.events.ForcedTransactionAddedEvent
import linea.contract.l1.LineaRollupFinalizedState
import linea.contract.l1.LineaRollupSmartContractClientReadOnlyFinalizedStateProvider
import linea.contrat.events.FactoryFinalizedStateUpdatedEvent
import linea.contrat.events.FactoryForcedTransactionAddedEvent
import linea.domain.BlockParameter
import linea.domain.EthLog
import linea.ethapi.FakeEthApiClient
import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import linea.forcedtx.ForcedTransactionRequest
import linea.forcedtx.ForcedTransactionResponse
import linea.forcedtx.ForcedTransactionsClient
import linea.log4j.configureLoggers
import linea.persistence.ftx.ForcedTransactionsDao
import net.consensys.zkevm.domain.ForcedTransactionRecord
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.CopyOnWriteArrayList
import kotlin.time.Clock
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

class FakeLineaRollupSmartContractClientReadOnlyFinalizedStateProvider(
  var l1FinalizedState: LineaRollupFinalizedState = LineaRollupFinalizedState(
    blockNumber = 0UL,
    blockTimestamp = Clock.System.now(),
    messageNumber = 0UL,
    forcedTransactionNumber = 10UL,
  ),
) :
  LineaRollupSmartContractClientReadOnlyFinalizedStateProvider {

  override fun getLatestFinalizedState(blockParameter: BlockParameter): SafeFuture<LineaRollupFinalizedState> {
    return SafeFuture.completedFuture(l1FinalizedState)
  }
}

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
    println("lineaSendForcedRawTransaction: ${transactions.map { it.ftxNumber }}")
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

class FakeForcedTransactionsDao(
  var records: MutableMap<ULong, ForcedTransactionRecord> = ConcurrentHashMap(),
) : ForcedTransactionsDao {

  override fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit> {
    records[ftx.ftxNumber] = ftx
    return SafeFuture.completedFuture(Unit)
  }

  override fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?> {
    return SafeFuture.completedFuture(records[ftxNumber])
  }

  override fun list(): SafeFuture<List<ForcedTransactionRecord>> {
    return SafeFuture.completedFuture(records.values.toList().sortedBy { it.ftxNumber })
  }

  override fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int> {
    var count = 0
    records.keys.forEach {
      if (it <= ftxNumber) {
        count++
        records.remove(it)
      }
    }
    return SafeFuture.completedFuture(count)
  }
}

@ExtendWith(VertxExtension::class)
class ForcedTransactionsAppTest {
  private val L1_CONTRACT_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01"
  private lateinit var l1Client: FakeEthApiClient
  private lateinit var vertx: Vertx
  private lateinit var fakeFinalStateProvider: FakeLineaRollupSmartContractClientReadOnlyFinalizedStateProvider
  private lateinit var ftxClient: FakeForcedTransactionsClient
  private lateinit var fxtDao: ForcedTransactionsDao

  @BeforeEach
  fun setUp(vertx: Vertx) {
    this.vertx = vertx
    l1Client = FakeEthApiClient(
      initialLogsDb = emptySet(),
      topicsTranslation = mapOf(
        FinalizedStateUpdatedEvent.topic to "FinalizedStateUpdated",
        ForcedTransactionAddedEvent.topic to "ForcedTransactionAdded",
      ),
      log = LogManager.getLogger("FakeEthApiClient"),
    )
    configureLoggers(
      rootLevel = Level.INFO,
      "FakeEthApiClient" to Level.INFO,
      "linea.ethapi" to Level.DEBUG,
      "linea.ftx" to Level.TRACE,
    )

    this.fakeFinalStateProvider = FakeLineaRollupSmartContractClientReadOnlyFinalizedStateProvider()
    this.ftxClient = FakeForcedTransactionsClient()
    this.fxtDao = FakeForcedTransactionsDao()
  }

  private fun createApp(
    l1PollingInterval: Duration = 100.milliseconds,
    l1EventSearchBlockChunk: UInt = 1000u,
    ftxSequencerSendingInterval: Duration = 100.milliseconds,
  ): ForcedTransactionsApp {
    val config = ForcedTransactionsApp.Config(
      l1PollingInterval = l1PollingInterval,
      l1ContractAddress = L1_CONTRACT_ADDRESS,
      l1HighestBlockTag = BlockParameter.Tag.FINALIZED,
      l1EventSearchBlockChunk = l1EventSearchBlockChunk,
      ftxSequencerSendingInterval = ftxSequencerSendingInterval,
    )

    return ForcedTransactionsApp(
      config = config,
      vertx = vertx,
      l1EthApiClient = l1Client,
      finalizedStateProvider = this.fakeFinalStateProvider,
      ftxClient = this.ftxClient,
      ftxDao = this.fxtDao,
    )
  }

  fun createFtxAddedEvent(
    l1BlockNumber: ULong,
    ftxNumber: ULong,
    l2DeadLine: ULong,
  ): EthLog {
    return FactoryForcedTransactionAddedEvent.createEthLog(
      l1BlockNumber = l1BlockNumber,
      contractAddress = L1_CONTRACT_ADDRESS,
      forcedTransactionNumber = ftxNumber,
      blockNumberDeadline = l2DeadLine,
    )
  }

  @Test
  fun `should send ftx to the sequencer in order`() {
    val finalizedStateEvent =
      FactoryFinalizedStateUpdatedEvent.createEthLog(
        blockNumber = 1_000UL,
        contractAddress = L1_CONTRACT_ADDRESS,
        l2FinalizedBlockNumber = 100UL,
        forcedTransactionNumber = 10UL,
      )
    val ftxAddedEvents = listOf(
      createFtxAddedEvent(
        l1BlockNumber = 900UL,
        ftxNumber = 9UL,
        l2DeadLine = 90UL,
      ),
      createFtxAddedEvent(
        l1BlockNumber = 990UL,
        ftxNumber = 10UL,
        l2DeadLine = 100UL,
      ),
      createFtxAddedEvent(
        l1BlockNumber = 1_100UL,
        ftxNumber = 11UL,
        l2DeadLine = 200UL,
      ),
      createFtxAddedEvent(
        l1BlockNumber = 1_200UL,
        ftxNumber = 12UL,
        l2DeadLine = 220UL,
      ),
    )
    l1Client.setLogs(listOf(finalizedStateEvent) + ftxAddedEvents)

    val l1PollingInterval = 50.milliseconds
    val app = createApp(
      l1PollingInterval = l1PollingInterval,
      l1EventSearchBlockChunk = 1000u,
    )
    this.l1Client.setFinalizedBlockTag(5_000UL)
    this.l1Client.setLatestBlockTag(10_000UL)
    this.ftxClient.setFtxInclusionResultAfterReception(
      ftxNumber = 11UL,
      l2BlockNumber = 2_000UL,
      inclusionResult = ForcedTransactionInclusionResult.Included,
    )
    app.start().get()

    await()
      .atMost(2.seconds.toJavaDuration())
      .untilAsserted {
        // ftx 12 is expected to be duplicated because sequencer does not have inclusion status
        // so coordinator shall resend it in case sequencer restarts
        assertThat(ftxClient.ftxReceivedIds).startsWith(11UL, 12UL, 12UL)
      }
    assertThat(fxtDao.list().get()).isEqualTo(
      listOf(
        ForcedTransactionRecord(
          ftxNumber = 11UL,
          inclusionResult = ForcedTransactionInclusionResult.Included,
          simulatedExecutionBlockNumber = this.ftxClient.ftxInclusionResults[11UL]!!.blockNumber,
          simulatedExecutionBlockTimestamp = this.ftxClient.ftxInclusionResults[11UL]!!.blockTimestamp,
        ),
      ),
    )
  }

  //
  // should send handle duplicated ftx with different submissions order
  // should be resilient to coordinator/sequencer restarts ensure submission order without gaps
}
