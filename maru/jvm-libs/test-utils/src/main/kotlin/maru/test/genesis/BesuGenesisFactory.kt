/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.test.genesis

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ObjectNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import maru.consensus.DifficultyAwareQbftConfig
import maru.consensus.ElFork
import maru.consensus.ForksSchedule
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.genesis.GenesisConfigurationFactory

class BesuGenesisFactory(
  val genesisTemplate: String = genesisTemplateLondonWithoutConsensus,
  val cliqueBlockTimeSeconds: UInt = 1u,
  val cliqueEmptyBlocks: Boolean = true,
) {
  private var forkSchedule: ForksSchedule? = null

  // we can cache the genesis if for the same fork schedule
  // which does not change per cluster (for now at least)
  // building the genesis for many nodes repeatedly is wasteful
  private var cachedGenesis: String? = null

  fun setForkSchedule(value: ForksSchedule) {
    this.forkSchedule = value
    // invalidate cached genesis when fork schedule changes
    cachedGenesis = null
  }

  fun create(): String {
    if (this.forkSchedule == null) {
      throw IllegalStateException("forkSchedule must be initialized before calling creating genesis")
    }

    return if (cachedGenesis != null) {
      return cachedGenesis!!
    } else {
      cachedGenesis =
        createGenesisWithClique(
          genesisTemplate,
          cliqueBlockTimeSeconds,
          cliqueEmptyBlocks,
          forkSchedule!!,
        )
      cachedGenesis!!
    }
  }

  companion object {
    val genesisTemplateLondonWithoutConsensus: String =
      GenesisConfigurationFactory.readGenesisFile("/besu-genesis-template.json")
    private val jsonObjectMapper: ObjectMapper = jacksonObjectMapper()

    fun createGenesisWithClique(
      genesisTemplate: String = genesisTemplateLondonWithoutConsensus,
      chainId: ULong,
      cliqueBlockTimeSeconds: UInt,
      cliqueEmptyBlocks: Boolean = true,
    ): String {
      require(cliqueBlockTimeSeconds in 1u..60u) { "cliqueBlockTimeSeconds must be between 1 and 60 seconds" }
      var updatedGenesis = genesisTemplate
      updatedGenesis = setGenesisConfigProperty(updatedGenesis, "chainId", chainId)
      updatedGenesis = setGenesisCliqueOptions(updatedGenesis, cliqueBlockTimeSeconds, cliqueEmptyBlocks)
      return updatedGenesis
    }

    fun createGenesisWithClique(
      genesisTemplate: String = genesisTemplateLondonWithoutConsensus,
      cliqueBlockTimeSeconds: UInt,
      cliqueEmptyBlocks: Boolean = true,
      forks: ForksSchedule,
    ): String {
      val terminalTotalDifficulty: ULong =
        forks.forks
          .firstOrNull { it.configuration is DifficultyAwareQbftConfig }
          ?.let {
            (it.configuration as DifficultyAwareQbftConfig).terminalTotalDifficulty
          } ?: 0UL

      var shanghaiTimestamp: ULong? = null
      var cancunTimestamp: ULong? = null
      var pragueTimestamp: ULong? = null
      forks.forks.forEach { forkSpec ->
        when (forkSpec.configuration.fork.elFork) {
          ElFork.Paris -> {} // nothing to do, terminalTotalDifficulty already set
          ElFork.Shanghai -> {
            shanghaiTimestamp = forkSpec.timestampSeconds
          }

          ElFork.Cancun -> {
            shanghaiTimestamp = shanghaiTimestamp ?: forkSpec.timestampSeconds
            cancunTimestamp = forkSpec.timestampSeconds
          }

          ElFork.Prague -> {
            shanghaiTimestamp = shanghaiTimestamp ?: forkSpec.timestampSeconds
            cancunTimestamp = cancunTimestamp ?: forkSpec.timestampSeconds
            pragueTimestamp = forkSpec.timestampSeconds
          }
        }
      }

      // Additional fork-specific genesis updates can be added here
      return createGenesisWithClique(
        genesisTemplate,
        forks.chainId.toULong(),
        cliqueBlockTimeSeconds,
        cliqueEmptyBlocks,
        terminalTotalDifficulty,
        shanghaiTimestamp,
        cancunTimestamp,
        pragueTimestamp,
      )
    }

    fun createGenesisWithClique(
      genesisTemplate: String = genesisTemplateLondonWithoutConsensus,
      chainId: ULong,
      cliqueBlockTimeSeconds: UInt,
      cliqueEmptyBlocks: Boolean = true,
      terminalTotalDifficulty: ULong? = null,
      shanghaiTimestamp: ULong? = null,
      cancunTimestamp: ULong? = null,
      pragueTimestamp: ULong? = null,
    ): String {
      var updatedGenesis = createGenesisWithClique(genesisTemplate, chainId, cliqueBlockTimeSeconds, cliqueEmptyBlocks)

      if (terminalTotalDifficulty != null) {
        updatedGenesis = setGenesisConfigProperty(updatedGenesis, "terminalTotalDifficulty", terminalTotalDifficulty)
      }
      if (shanghaiTimestamp != null) {
        require(
          terminalTotalDifficulty != null,
        ) { "terminalTotalDifficulty must be defined when shanghaiTimestamp is defined" }
        updatedGenesis = setGenesisConfigProperty(updatedGenesis, "shanghaiTime", shanghaiTimestamp)
      }
      if (cancunTimestamp != null) {
        updatedGenesis = setGenesisConfigProperty(updatedGenesis, "cancunTime", cancunTimestamp)
      }
      if (pragueTimestamp != null) {
        updatedGenesis = setGenesisConfigProperty(updatedGenesis, "pragueTime", pragueTimestamp)
      }
      return updatedGenesis
    }

    private fun setGenesisCliqueOptions(
      genesis: String,
      blockTimeSeconds: UInt,
      createEmptyBlocks: Boolean,
    ): String {
      val rootNode = jsonObjectMapper.readTree(genesis) as ObjectNode
      val configNode = rootNode.get("config") as ObjectNode

      // Create or replace the clique configuration
      val cliqueNode = jsonObjectMapper.createObjectNode()
      cliqueNode.put("blockperiodseconds", blockTimeSeconds.toLong())
      cliqueNode.put("epochlength", blockTimeSeconds.toLong())
      cliqueNode.put("createemptyblocks", createEmptyBlocks)

      configNode.set<ObjectNode>("clique", cliqueNode)

      return jsonObjectMapper.writeValueAsString(rootNode)
    }

    private fun setGenesisConfigProperty(
      genesis: String,
      key: String,
      value: ULong,
    ): String {
      val rootNode = jsonObjectMapper.readTree(genesis) as ObjectNode
      val configNode = rootNode.get("config") as ObjectNode
      configNode.put(key, value.toLong())
      return jsonObjectMapper.writeValueAsString(rootNode)
    }
  }
}
