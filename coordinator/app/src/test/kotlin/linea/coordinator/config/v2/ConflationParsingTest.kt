package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.ConflationToml
import linea.coordinator.config.v2.toml.parseConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.net.URI
import kotlin.time.Duration.Companion.seconds

class ConflationParsingTest {
  companion object {
    val toml = """
      [conflation]
      disabled = true
      blocks-limit = 2
      conflation-calculator-version = "1.0.0"
      conflation-deadline = "PT6S" # =3*l2_block_time
      conflation-deadline-check-interval = "PT3S"
      conflation-deadline-last-block-confirmation-delay = "PT2S" # recommended: at least 2 * blockInterval
      l2-fetch-blocks-limit = 4_000
      l2-block-creation-endpoint = "http://sequencer:8545"
      l2-logs-endpoint = "http://sequencer:8545"
      consistent-number-of-blocks-on-l1-to-wait = 1

      [conflation.blob-compression]
      blob-size-limit = 102_400 # 100KB
      handler-polling-interval = "PT1S"
      # default batches limit is  aggregation-proofs-limit -1
      # batches-limit must be less than or equal to aggregation-proofs-limit-1
      batches-limit = 1

      [conflation.proof-aggregation]
      proofs-limit = 3
      deadline = "PT1M"
      coordinator-polling-interval = "PT2S"
      deadline-check-interval = "PT8S"
      target-end-blocks = [10, 20, 30_000]
    """.trimIndent()
    val config = ConflationToml(
      disabled = true,
      blocksLimit = 2,
      conflationCalculatorVersion = "1.0.0",
      conflationDeadline = 6.seconds,
      l2FetchBlocksLimit = 4000u,
      l2BlockCreationEndpoint = URI.create("http://sequencer:8545").toURL(),
      l2LogsEndpoint = URI.create("http://sequencer:8545").toURL(),
      consistentNumberOfBlocksOnL1ToWait = 1,
      blobCompression = ConflationToml.BlobCompressionToml(
        blobSizeLimit = 102_400UL,
        handlerPollingInterval = 1.seconds,
        batchesLimit = 1u
      ),
      proofAggregation = ConflationToml.ProofAggregationToml(
        proofsLimit = 3uL,
        deadline = 60.seconds,
        coordinatorPollingInterval = 2.seconds,
        deadlineCheckInterval = 8.seconds,
        targetEndBlocks = listOf(10uL, 20uL, 30_000uL)
      )
    )
  }

  data class WrapperConfig(
    val conflation: ConflationToml
  )

  @Test
  fun `should parse conflation toml configs - full`() {
    assertThat(
      parseConfig<WrapperConfig>(toml).conflation
    )
      .isEqualTo(config)
  }

  @Test
  fun `should parse conflation toml configs and provide defaults`() {
    val toml = """
      [conflation]
      conflation-calculator-version = "1.0.0"

      [conflation.blob-compression]
      blob-size-limit = 102_400 # 100KB

      [conflation.proof-aggregation]
      proofs-limit = 3
    """.trimIndent()
    assertThat(
      parseConfig<WrapperConfig>(toml).conflation
    ).isEqualTo(
      ConflationToml(
        disabled = false,
        blocksLimit = null,
        conflationCalculatorVersion = "1.0.0",
        conflationDeadline = null,
        l2FetchBlocksLimit = null,
        l2BlockCreationEndpoint = null,
        l2LogsEndpoint = null,
        consistentNumberOfBlocksOnL1ToWait = 32,
        blobCompression = ConflationToml.BlobCompressionToml(
          blobSizeLimit = 102_400UL,
          handlerPollingInterval = 1.seconds,
          batchesLimit = null
        ),
        proofAggregation = ConflationToml.ProofAggregationToml(
          proofsLimit = 3uL,
          deadline = null,
          deadlineCheckInterval = 30.seconds,
          coordinatorPollingInterval = 3.seconds,
          targetEndBlocks = null
        )
      )
    )
  }
}
