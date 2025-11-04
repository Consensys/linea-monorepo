/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import picocli.CommandLine

class MaruAppCliTest {
  private val cmd = CommandLine(MaruAppCli(true))

  @BeforeEach
  fun setUp() {
    cmd.registerConverter(
      Network::class.java,
      KebabToEnumConverter(Network::class.java),
    )
  }

  @Test
  fun `should parse commandline args with 'network' as linea-mainnet`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--network=linea-mainnet",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(0)

    val cli = cmd.getCommand<MaruAppCli>()
    assertThat(cli.genesisOptions!!.network!!.networkNameInKebab).isEqualTo("linea-mainnet")
    assertThat(cli.genesisOptions!!.genesisFile!!).isEqualTo(buildInGenesisFileResourcePath("linea-mainnet"))
    assertThat(cli.configFiles!!.first().path).isEqualTo("./maru.config.toml")
  }

  @Test
  fun `should parse commandline args with network as linea-seoplia with case-insensitive`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--network=LINEA-SEPOLIA",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(0)

    val cli = cmd.getCommand<MaruAppCli>()
    assertThat(cli.genesisOptions!!.network!!.networkNameInKebab).isEqualTo("linea-sepolia")
    assertThat(cli.genesisOptions!!.genesisFile!!).isEqualTo(buildInGenesisFileResourcePath("linea-sepolia"))
    assertThat(cli.configFiles!!.first().path).isEqualTo("./maru.config.toml")
  }

  @Test
  fun `should parse commandline args with 'maru-genesis-file' specified`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--maru-genesis-file=./maru.genesis.json",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(0)
    val cli = cmd.getCommand<MaruAppCli>()
    assertThat(cli.genesisOptions!!.network).isNull()
    assertThat(cli.genesisOptions!!.genesisFile!!).isEqualTo("./maru.genesis.json")
    assertThat(cli.configFiles!!.first().path).isEqualTo("./maru.config.toml")
  }

  @Test
  fun `should parse commandline args with 'genesis-file' specified`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--genesis-file=./maru.genesis.json",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(0)
    val cli = cmd.getCommand<MaruAppCli>()
    assertThat(cli.genesisOptions!!.network).isNull()
    assertThat(cli.genesisOptions!!.genesisFile!!).isEqualTo("./maru.genesis.json")
    assertThat(cli.configFiles!!.first().path).isEqualTo("./maru.config.toml")
  }

  @Test
  fun `should parse commandline args and default to linea-mainnet with only 'config' specified`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(0)
    val cli = cmd.getCommand<MaruAppCli>()
    assertThat(cli.genesisOptions!!.network!!.networkNameInKebab).isEqualTo("linea-mainnet")
    assertThat(cli.genesisOptions!!.genesisFile!!).isEqualTo(buildInGenesisFileResourcePath("linea-mainnet"))
    assertThat(cli.configFiles!!.first().path).isEqualTo("./maru.config.toml")
  }

  @Test
  fun `should fail to parse commandline args with both 'genesis-file' and 'network' specified`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--genesis-file=./maru.genesis.json",
        "--network=linea-mainnet",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(2)
  }

  @Test
  fun `should fail to parse commandline args without 'config' specified`() {
    val args =
      listOf(
        "--network=linea-mainnet",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(2)
  }

  @Test
  fun `should fail to parse commandline args with 'network' as both linea-mainnet and linea-sepolia`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--network=linea-mainnet",
        "--network=linea-sepolia",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(2)
  }

  @Test
  fun `should fail to parse commandline args with invalid 'network'`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--network=linea-testnet",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(2)
  }

  @Test
  fun `should fail to parse commandline args with 'genesis-file' and 'maru-genesis-file' both specified`() {
    val args =
      listOf(
        "--config=./maru.config.toml",
        "--genesis-file=./maru.genesis.json",
        "--maru-genesis-file=./maru.genesis.json",
      )
    val exitCode = cmd.execute(*args.toTypedArray())
    assertThat(exitCode).isEqualTo(2)
  }
}
