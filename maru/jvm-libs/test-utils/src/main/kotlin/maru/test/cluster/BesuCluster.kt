/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.test.cluster

import java.net.URI
import org.apache.logging.log4j.LogManager
import org.hyperledger.besu.tests.acceptance.dsl.condition.Condition
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.RunnableNode
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfiguration
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions

class BesuCluster(
  private val clusterConfiguration: ClusterConfiguration = ClusterConfigurationBuilder().build(),
  private val net: NetConditions = NetConditions(NetTransactions()),
  private val besuNodeRunner: BesuNodeRunner = ThreadBesuNodeRunner(),
) : AutoCloseable {
  val log = LogManager.getLogger(BesuCluster::class.java)
  internal val nodes: MutableMap<String, BesuNode> = HashMap()
  internal val bootnodes: MutableList<URI> = ArrayList()

  private fun startCluster(awaitPeerDiscovery: Boolean) {
    require(nodes.isNotEmpty()) { "Can't start a cluster with no nodes" }
    val nodesList = nodes.values.toList()
    val bootnode = selectAndStartBootnode(nodesList)

    nodesList
      .parallelStream()
      .filter { node -> bootnode?.let { it != node } ?: true }
      .peek { node -> log.info("starting non-bootnode {}", node.name) }
      .forEach { startNode(it) }

    if (awaitPeerDiscovery) {
      val timeoutSeconds = clusterConfiguration.peerDiscoveryTimeoutSeconds
      for (node in nodesList) {
        log.info(
          "Awaiting peer discovery for node {}, expecting {} peers, timeout {} seconds",
          node.name,
          nodes.size - 1,
          timeoutSeconds,
        )
        node.awaitPeerDiscovery(net.awaitPeerCount(nodes.size - 1, timeoutSeconds))
      }
    }
    log.info("Cluster startup complete.")
  }

  private fun selectAndStartBootnode(nodes: List<BesuNode>): RunnableNode? {
    val bootnode = nodes.firstOrNull(::isBootnodeEligible)

    bootnode?.let { b ->
      log.info("Selected node {} as bootnode", b.name)
      startNode(b, true)
      bootnodes.add(b.enodeUrl())
    }

    return bootnode
  }

  fun isBootnodeEligible(node: RunnableNode): Boolean =
    node.configuration.isBootnodeEligible &&
      node.configuration.isP2pEnabled &&
      node.configuration.isDiscoveryEnabled

  fun addNode(node: BesuNode) {
    require(nodes[node.name] == null) { "Node with name ${node.name} already exists" }
    nodes[node.name] = node
  }

  fun addNodeAndStart(
    node: BesuNode,
    awaitPeerDiscovery: Boolean = false,
  ) {
    addNode(node)
    if (bootnodes.isEmpty()) {
      selectAndStartBootnode(listOf(node))
    } else {
      startNode(node)
    }

    if (awaitPeerDiscovery) {
      node.awaitPeerDiscovery(net.awaitPeerCount(nodes.size - 1))
    }
  }

  private fun runNodeStart(node: RunnableNode) {
    log.info(
      "Starting node {} (id = {}...{})",
      node.name,
      node.nodeId.substring(0, 4),
      node.nodeId.substring(124),
    )
    node.start(besuNodeRunner)
  }

  private fun startNode(
    node: RunnableNode,
    isBootNode: Boolean = false,
  ) {
    node.configuration.bootnodes = if (isBootNode) emptyList() else bootnodes

    if (node.configuration.genesisConfig.isEmpty) {
      node.configuration.genesisConfigProvider
        .create(nodes.values)
        .ifPresent { node.configuration.setGenesisConfig(it) }
    }
    runNodeStart(node)
  }

  fun start(awaitPeerDiscovery: Boolean = true) = startCluster(awaitPeerDiscovery)

  fun stop() {
    // to avoid ConcurrentModificationException
    val nodesList = nodes.values.toList()
    for (node in nodesList) {
      stopNode(node)
    }
  }

  fun stopNode(nodeName: String) {
    val node = nodes[nodeName] ?: throw IllegalArgumentException("Node $nodeName not found in cluster")
    stopNode(node)
  }

  private fun stopNode(node: BesuNode) {
    besuNodeRunner.stopNode(node)
    node.stop()
    nodes.remove(node.name)
  }

  override fun close() {
    stop()
    for (node in nodes.values) {
      node.close()
    }
    besuNodeRunner.shutdown()
  }

  fun verify(expected: Condition) {
    if (nodes.isEmpty()) {
      throw IllegalStateException("Attempt to verify an empty cluster")
    }
    for (node in nodes.values) {
      expected.verify(node)
    }
  }

  fun verifyOnActiveNodes(condition: Condition) {
    nodes.values
      .filter { node -> besuNodeRunner.isActive(node.name) }
      .forEach { condition.verify(it) }
  }

  fun startConsoleCapture() {
    besuNodeRunner.startConsoleCapture()
  }

  fun getConsoleContents(): String = besuNodeRunner.consoleContents

  // Expose besuNodeRunner for compatibility
  val nodeRunner: BesuNodeRunner get() = besuNodeRunner
}
