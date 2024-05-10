package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.L2MessageService
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L2AccountManager
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.protocol.Web3j
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class L2QuerierIntegrationTest {
  private lateinit var testL2MessageManagerContractAddress: String
  private lateinit var l2Contract: L2MessageService
  private lateinit var l2Web3jClient: Web3j

  @BeforeEach
  fun beforeEach(
    vertx: Vertx
  ) {
    val deploymentResult = ContractsManager.get().deployL2MessageService().get()
    testL2MessageManagerContractAddress = deploymentResult.contractAddress
    l2Contract = ContractsManager.get().connectL2MessageService(
      contractAddress = deploymentResult.contractAddress,
      transactionManager = deploymentResult.anchorerOperator.txManager
    )
    l2Web3jClient = L2AccountManager.web3jClient
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `anchored hashes are returned correctly`(vertx: Vertx, testContext: VertxTestContext) {
    val config = L2QuerierImpl.Config(0u, 2u, 2u, testL2MessageManagerContractAddress)
    val l2Querier = L2QuerierImpl(l2Web3jClient, l2Contract, config, vertx)
    val events = createRandomSendMessageEvents(10UL)

    val l2MessageAnchorerImpl = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      L2MessageAnchorerImpl.Config(12.seconds, 8u, 0)
    )

    l2MessageAnchorerImpl.anchorMessages(events, Bytes32.random().toArray())
      .thenCompose { _ ->
        l2Querier.findLastFinalizedAnchoredEvent().thenCompose {
          l2Querier.getMessageHashStatus(it!!.messageHash).thenApply {
            testContext.verify {
              Assertions.assertThat(it).isNotNull
              Assertions.assertThat(it).isEqualTo(BigInteger.valueOf(1))
            }.completeNow()
          }
        }.whenException(testContext::failNow)
      }
  }
}
