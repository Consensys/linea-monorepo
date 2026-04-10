package net.consensys.zkevm.coordinator.blockcreation

import linea.domain.createBlock
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.util.concurrent.atomic.AtomicLong
import kotlin.time.Instant

class ConflationTargetCheckpointPauseControllerTest {
  private class TestL1Provider(
    initial: Long = 0L,
  ) : LatestL1FinalizedBlockProviderSync {
    private val inner = AtomicLong(initial)
    override fun getLatestL1FinalizedBlock(): Long = inner.get()
    fun setLatest(block: Long) {
      inner.set(block)
    }
  }

  private fun pauseConfig(
    initialTs: Instant = Instant.fromEpochSeconds(0L),
    targetEndBlocks: Set<ULong> = emptySet(),
    targetTimestamps: List<Instant> = emptyList(),
    waitL1: Boolean = false,
    waitApi: Boolean = false,
  ): ConflationTargetCheckpointPauseController.Config =
    ConflationTargetCheckpointPauseController.Config(
      initialLastImportedBlockTimestamp = initialTs,
      targetEndBlocks = targetEndBlocks,
      targetTimestamps = targetTimestamps,
      waitTargetBlockL1Finalization = waitL1,
      waitApiResumeAfterTargetBlock = waitApi,
    )

  @Test
  fun `does nothing when both wait switches are false`() {
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetEndBlocks = setOf(9uL),
        ),
        l1,
      )
    assertThat(c.pauseFeatureEnabled).isFalse()
    c.importBlock(createBlock(number = 10uL))
    assertThat(c.isPausedForTests()).isFalse()
  }

  @Test
  fun `pauses at target block number plus one and resumes on L1 only`() {
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetEndBlocks = setOf(9uL),
          waitL1 = true,
        ),
        l1,
      )
    c.importBlock(createBlock(number = 10uL))
    assertThat(c.isPausedForTests()).isTrue()
    assertThat(c.shouldPauseConflation()).isTrue()
    l1.setLatest(8L)
    assertThat(c.shouldPauseConflation()).isTrue()
    l1.setLatest(9L)
    assertThat(c.shouldPauseConflation()).isFalse()
    assertThat(c.isPausedForTests()).isFalse()
  }

  @Test
  fun `timestamp pauses only on first block that crosses from strictly below threshold`() {
    val tbt = Instant.fromEpochSeconds(1000L)
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetTimestamps = listOf(tbt),
          waitL1 = true,
        ),
        l1,
      )
    c.importBlock(createBlock(number = 1uL, timestamp = Instant.fromEpochSeconds(999L)))
    assertThat(c.isPausedForTests()).isFalse()
    c.importBlock(createBlock(number = 2uL, timestamp = tbt))
    assertThat(c.isPausedForTests()).isTrue()
  }

  @Test
  fun `after timestamp resume later blocks above threshold do not pause again for same list`() {
    val tbt = Instant.fromEpochSeconds(1000L)
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetTimestamps = listOf(tbt),
          waitL1 = true,
        ),
        l1,
      )
    c.importBlock(createBlock(number = 1uL, timestamp = Instant.fromEpochSeconds(999L)))
    c.importBlock(createBlock(number = 2uL, timestamp = tbt))
    assertThat(c.isPausedForTests()).isTrue()
    l1.setLatest(1L)
    assertThat(c.shouldPauseConflation()).isFalse()
    assertThat(c.isPausedForTests()).isFalse()
    c.importBlock(createBlock(number = 3uL, timestamp = Instant.fromEpochSeconds(10_000L)))
    assertThat(c.isPausedForTests()).isFalse()
  }

  @Test
  fun `pauses on timestamp threshold and resumes when L1 reaches target block timestamp by block number minus one`() {
    val tbt = Instant.fromEpochSeconds(1000L)
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetTimestamps = listOf(tbt),
          waitL1 = true,
        ),
        l1,
      )
    c.importBlock(
      createBlock(number = 5uL, timestamp = tbt),
    )
    assertThat(c.isPausedForTests()).isTrue()
    l1.setLatest(3L)
    assertThat(c.shouldPauseConflation()).isTrue()
    l1.setLatest(4L)
    assertThat(c.shouldPauseConflation()).isFalse()
    assertThat(c.isPausedForTests()).isFalse()
  }

  @Test
  fun `API resume when waitApi true`() {
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetEndBlocks = setOf(9uL),
          waitApi = true,
        ),
        l1,
      )
    c.importBlock(createBlock(number = 10uL))
    assertThat(c.isPausedForTests()).isTrue()
    l1.setLatest(100L)
    assertThat(c.shouldPauseConflation()).isTrue()
    assertThat(c.signalResumeFromApi()).isTrue()
    assertThat(c.shouldPauseConflation()).isFalse()
    assertThat(c.isPausedForTests()).isFalse()
  }

  @Test
  fun `signalResumeFromApi returns false when API gate disabled`() {
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          waitL1 = true,
        ),
        l1,
      )
    assertThat(c.signalResumeFromApi()).isFalse()
  }

  @Test
  fun `signalResumeFromApi returns false when no checkpoint pause active`() {
    val l1 = TestL1Provider()
    val c =
      ConflationTargetCheckpointPauseController(
        pauseConfig(
          targetEndBlocks = setOf(9uL),
          waitApi = true,
        ),
        l1,
      )
    assertThat(c.signalResumeFromApi()).isFalse()
  }
}
