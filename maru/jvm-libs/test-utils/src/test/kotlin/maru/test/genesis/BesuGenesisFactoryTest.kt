/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.test.genesis

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import kotlin.random.Random
import kotlinx.datetime.Instant
import maru.consensus.ChainFork
import maru.consensus.ClFork
import maru.consensus.DifficultyAwareQbftConfig
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.QbftConsensusConfig
import maru.core.Validator
import org.assertj.core.api.Assertions
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test

class BesuGenesisFactoryTest {
  val objectMapper = jacksonObjectMapper()

  @Nested
  inner class CreateGenesisWithCliqueBasic {
    @Test
    fun `should create genesis with just clique configuration - empty blocks true`() {
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 13UL,
          cliqueBlockTimeSeconds = 4U,
          cliqueEmptyBlocks = true,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 4UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 4UL)
      Assertions.assertThat(config.get("clique").get("createemptyblocks").isBoolean).isTrue()
      Assertions.assertThat(config.get("clique").get("createemptyblocks").asBoolean()).isEqualTo(true)
      Assertions.assertThat(config.get("terminalTotalDifficulty")).isNull()

      Assertions.assertThat(config.get("shanghaiTime")).isNull()
      Assertions.assertThat(config.get("cancunTime")).isNull()
      Assertions.assertThat(config.get("pragueTime")).isNull()
    }

    @Test
    fun `should handle empty blocks false`() {
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 1337UL,
          cliqueBlockTimeSeconds = 2U,
          cliqueEmptyBlocks = false,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      Assertions.assertThat(config.get("clique").get("createemptyblocks").asBoolean()).isEqualTo(false)
    }
  }

  @Nested
  inner class CreateGenesisWithCliqueAndForks {
    val validators = setOf(Validator(address = Random.nextBytes(20)))

    @Test
    fun `should create genesis with all fork timestamps`() {
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 13UL,
          cliqueBlockTimeSeconds = 2U,
          cliqueEmptyBlocks = true,
          terminalTotalDifficulty = 500UL,
          shanghaiTimestamp = 1000UL,
          cancunTimestamp = 2000UL,
          pragueTimestamp = 3000UL,
          osakaTimestamp = 4000UL,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 2UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 2UL)
      Assertions.assertThat(config.get("clique").get("createemptyblocks").isBoolean).isTrue()
      Assertions.assertThat(config.get("clique").get("createemptyblocks").asBoolean()).isEqualTo(true)
      assertIsNumberWithValue(config.get("terminalTotalDifficulty"), 500UL)
      assertIsNumberWithValue(config.get("shanghaiTime"), 1000UL)
      assertIsNumberWithValue(config.get("cancunTime"), 2000UL)
      assertIsNumberWithValue(config.get("pragueTime"), 3000UL)
      assertIsNumberWithValue(config.get("osakaTime"), 4000UL)
    }

    @Test
    fun `should create genesis with forks schedule with all forks`() {
      val ttdForkSpec =
        ForkSpec(
          timestampSeconds = 0UL,
          blockTimeSeconds = 2U,
          configuration =
            DifficultyAwareQbftConfig(
              terminalTotalDifficulty = 2000UL,
              postTtdConfig =
                QbftConsensusConfig(
                  validatorSet = validators,
                  fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Paris),
                ),
            ),
        )
      val shanghaiForkSpec =
        ForkSpec(
          timestampSeconds =
            Instant.Companion
              .parse("2025-10-20T00:00:00Z")
              .epochSeconds
              .toULong(),
          blockTimeSeconds = 2U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Shanghai),
            ),
        )
      val cancunForkSpec =
        ForkSpec(
          timestampSeconds =
            Instant.Companion
              .parse("2025-10-21T00:00:00Z")
              .epochSeconds
              .toULong(),
          blockTimeSeconds = 2U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Cancun),
            ),
        )
      val pragueForkSpec =
        ForkSpec(
          timestampSeconds =
            Instant.Companion
              .parse("2025-10-22T00:00:00Z")
              .epochSeconds
              .toULong(),
          blockTimeSeconds = 2U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Prague),
            ),
        )

      val osakaForkSpec =
        ForkSpec(
          timestampSeconds =
            Instant.Companion
              .parse("2025-10-23T00:00:00Z")
              .epochSeconds
              .toULong(),
          blockTimeSeconds = 2U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Osaka),
            ),
        )

      val forksSchedule =
        ForksSchedule(13U, listOf(ttdForkSpec, shanghaiForkSpec, cancunForkSpec, pragueForkSpec, osakaForkSpec))

      val result =
        BesuGenesisFactory.createGenesisWithClique(
          cliqueBlockTimeSeconds = 4U,
          cliqueEmptyBlocks = true,
          forks = forksSchedule,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 4UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 4UL)
      assertIsBooleanWithValue(config.get("clique").get("createemptyblocks"), true)
      assertIsNumberWithValue(
        config.get("terminalTotalDifficulty"),
        (ttdForkSpec.configuration as DifficultyAwareQbftConfig).terminalTotalDifficulty,
      )
      assertIsNumberWithValue(config.get("shanghaiTime"), shanghaiForkSpec.timestampSeconds)
      assertIsNumberWithValue(config.get("cancunTime"), cancunForkSpec.timestampSeconds)
      assertIsNumberWithValue(config.get("pragueTime"), pragueForkSpec.timestampSeconds)
      assertIsNumberWithValue(config.get("osakaTime"), osakaForkSpec.timestampSeconds)
    }

    @Test
    fun `should create genesis with forks schedule from prague`() {
      val pragueForkSpec =
        ForkSpec(
          timestampSeconds = 1000UL,
          blockTimeSeconds = 5U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Prague),
            ),
        )
      val osakaForkSpec =
        ForkSpec(
          timestampSeconds = 2000UL,
          blockTimeSeconds = 5U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Osaka),
            ),
        )
      val forksSchedule = ForksSchedule(13U, listOf(pragueForkSpec, osakaForkSpec))
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          cliqueBlockTimeSeconds = 4U,
          cliqueEmptyBlocks = true,
          forks = forksSchedule,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 4UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 4UL)
      assertIsBooleanWithValue(config.get("clique").get("createemptyblocks"), true)
      assertIsNumberWithValue(config.get("terminalTotalDifficulty"), 0UL)
      assertIsNumberWithValue(config.get("shanghaiTime"), 0UL)
      assertIsNumberWithValue(config.get("cancunTime"), 0UL)
      assertIsNumberWithValue(config.get("pragueTime"), 1000UL)
      assertIsNumberWithValue(config.get("osakaTime"), 2000UL)
    }

    @Test
    fun `should create genesis with fork schedule with interleaved forks`() {
      val cancunForkSpec =
        ForkSpec(
          timestampSeconds = 0UL,
          blockTimeSeconds = 5U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Cancun),
            ),
        )
      val osakaForkSpec =
        ForkSpec(
          timestampSeconds = 2000UL,
          blockTimeSeconds = 5U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Osaka),
            ),
        )
      val forksSchedule = ForksSchedule(13U, listOf(cancunForkSpec, osakaForkSpec))
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          cliqueBlockTimeSeconds = 4U,
          cliqueEmptyBlocks = true,
          forks = forksSchedule,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 4UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 4UL)
      assertIsBooleanWithValue(config.get("clique").get("createemptyblocks"), true)
      assertIsNumberWithValue(config.get("terminalTotalDifficulty"), 0UL)
      assertIsNumberWithValue(config.get("shanghaiTime"), 0UL)
      assertIsNumberWithValue(config.get("cancunTime"), 0UL)
      assertIsNumberWithValue(config.get("pragueTime"), 2000UL)
      assertIsNumberWithValue(config.get("osakaTime"), 2000UL)
    }
  }

  @Nested
  inner class EdgeCases {
    @Test
    fun `should require terminalTotalDifficulty when shanghaiTimestamp is provided`() {
      assertThatThrownBy {
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 1337UL,
          cliqueBlockTimeSeconds = 1U,
          shanghaiTimestamp = 500UL,
        )
      }.isInstanceOf(IllegalArgumentException::class.java)
        .hasMessageContaining("terminalTotalDifficulty must be defined when shanghaiTimestamp is defined")
    }

    @Test
    fun `should fail on zero block time`() {
      assertThatThrownBy {
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 1337UL,
          cliqueBlockTimeSeconds = 0U,
        )
      }.isInstanceOf(IllegalArgumentException::class.java)
        .hasMessageContaining("cliqueBlockTimeSeconds")
    }

    @Test
    fun `should preserve other config properties`() {
      val result =
        BesuGenesisFactory.createGenesisWithClique(
          chainId = 5555UL,
          cliqueBlockTimeSeconds = 7U,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      // Check that original properties are preserved
      Assertions.assertThat(config.get("homesteadBlock").asInt()).isEqualTo(0)
      Assertions.assertThat(config.get("eip150Block").asInt()).isEqualTo(0)
      // Check that new properties are added
      Assertions.assertThat(config.get("chainId").asText()).isEqualTo("5555")
      Assertions.assertThat(config.get("clique")).isNotNull
    }
  }
}
