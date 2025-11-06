/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.test.genesis

import com.fasterxml.jackson.databind.JsonNode
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
        BesuGenesisFactory.Companion.createGenesisWithClique(
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
        BesuGenesisFactory.Companion.createGenesisWithClique(
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
    @Test
    fun `should create genesis with all fork timestamps`() {
      val result =
        BesuGenesisFactory.Companion.createGenesisWithClique(
          chainId = 13UL,
          cliqueBlockTimeSeconds = 2U,
          cliqueEmptyBlocks = true,
          terminalTotalDifficulty = 500UL,
          shanghaiTimestamp = 1000UL,
          cancunTimestamp = 2000UL,
          pragueTimestamp = 3000UL,
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
    }

    @Test
    fun `should handle partial fork configuration`() {
      val result =
        BesuGenesisFactory.Companion.createGenesisWithClique(
          chainId = 1337UL,
          cliqueBlockTimeSeconds = 1U,
          terminalTotalDifficulty = 0UL,
          shanghaiTimestamp = 500UL,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("terminalTotalDifficulty"), 0UL)
      assertIsNumberWithValue(config.get("shanghaiTime"), 500UL)
      Assertions.assertThat(config.get("cancunTime")).isNull()
      Assertions.assertThat(config.get("pragueTime")).isNull()
    }

    @Test
    fun `should create genesis with forks schedule with all forks`() {
      val validators = setOf(Validator(address = Random.Default.nextBytes(20)))
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

      val forksSchedule = ForksSchedule(13U, listOf(ttdForkSpec, shanghaiForkSpec, cancunForkSpec, pragueForkSpec))

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
      Assertions.assertThat(config.get("clique").get("createemptyblocks").isBoolean).isTrue()
      Assertions.assertThat(config.get("clique").get("createemptyblocks").asBoolean()).isEqualTo(true)
      assertIsNumberWithValue(
        config.get("terminalTotalDifficulty"),
        (ttdForkSpec.configuration as DifficultyAwareQbftConfig).terminalTotalDifficulty,
      )
      assertIsNumberWithValue(config.get("shanghaiTime"), shanghaiForkSpec.timestampSeconds)
      assertIsNumberWithValue(config.get("cancunTime"), cancunForkSpec.timestampSeconds)
      assertIsNumberWithValue(config.get("pragueTime"), pragueForkSpec.timestampSeconds)
    }

    @Test
    fun `should create genesis with forks schedule from prague`() {
      val validators = setOf(Validator(address = Random.Default.nextBytes(20)))
      val forkSpec =
        ForkSpec(
          timestampSeconds = 1000UL,
          blockTimeSeconds = 5U,
          configuration =
            QbftConsensusConfig(
              validatorSet = validators,
              fork = ChainFork(ClFork.QBFT_PHASE1, ElFork.Prague),
            ),
        )
      val forksSchedule = ForksSchedule(13U, listOf(forkSpec))
      val result =
        BesuGenesisFactory.Companion.createGenesisWithClique(
          cliqueBlockTimeSeconds = 4U,
          cliqueEmptyBlocks = true,
          forks = forksSchedule,
        )

      val jsonNode = objectMapper.readTree(result)
      val config = jsonNode.get("config")

      assertIsNumberWithValue(config.get("chainId"), 13UL)
      assertIsNumberWithValue(config.get("clique").get("blockperiodseconds"), 4UL)
      assertIsNumberWithValue(config.get("clique").get("epochlength"), 4UL)
      Assertions.assertThat(config.get("clique").get("createemptyblocks").isBoolean).isTrue()
      Assertions.assertThat(config.get("clique").get("createemptyblocks").asBoolean()).isEqualTo(true)
      assertIsNumberWithValue(config.get("terminalTotalDifficulty"), 0UL)
      assertIsNumberWithValue(config.get("shanghaiTime"), 1000UL)
      assertIsNumberWithValue(config.get("cancunTime"), 1000UL)
      assertIsNumberWithValue(config.get("pragueTime"), 1000UL)
    }
  }

  fun assertIsNumberWithValue(
    node: JsonNode,
    expectedValue: ULong,
  ) {
    Assertions.assertThat(node.isNumber).isTrue
    Assertions.assertThat(node.asLong()).isEqualTo(expectedValue.toLong())
  }

  @Nested
  inner class EdgeCases {
    @Test
    fun `should require terminalTotalDifficulty when shanghaiTimestamp is provided`() {
      Assertions
        .assertThatThrownBy {
          BesuGenesisFactory.Companion.createGenesisWithClique(
            chainId = 1337UL,
            cliqueBlockTimeSeconds = 1U,
            shanghaiTimestamp = 500UL,
          )
        }.isInstanceOf(IllegalArgumentException::class.java)
        .hasMessageContaining("terminalTotalDifficulty must be defined when shanghaiTimestamp is defined")
    }

    @Test
    fun `should fail on zero block time`() {
      val result =
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
        BesuGenesisFactory.Companion.createGenesisWithClique(
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
