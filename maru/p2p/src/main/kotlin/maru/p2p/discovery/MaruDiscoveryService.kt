/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.discovery

import java.time.Duration
import java.util.concurrent.TimeUnit
import java.util.function.Consumer
import maru.config.P2P
import maru.consensus.ForkIdHashProvider
import net.consensys.linea.async.toSafeFuture
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.units.bigints.UInt64
import org.ethereum.beacon.discovery.DiscoverySystem
import org.ethereum.beacon.discovery.DiscoverySystemBuilder
import org.ethereum.beacon.discovery.schema.EnrField
import org.ethereum.beacon.discovery.schema.NodeRecord
import org.ethereum.beacon.discovery.schema.NodeRecordBuilder
import org.ethereum.beacon.discovery.schema.NodeRecordFactory
import org.hyperledger.besu.plugin.services.MetricsSystem
import tech.pegasys.teku.infrastructure.async.AsyncRunner
import tech.pegasys.teku.infrastructure.async.Cancellable
import tech.pegasys.teku.infrastructure.async.MetricTrackingExecutorFactory
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.async.ScheduledExecutorAsyncRunner
import tech.pegasys.teku.networking.p2p.discovery.discv5.SecretKeyParser

class MaruDiscoveryService(
  privateKeyBytes: ByteArray,
  private val p2pConfig: P2P,
  private val forkIdHashProvider: ForkIdHashProvider,
  metricsSystem: MetricsSystem,
) {
  companion object {
    val BOOTNODE_REFRESH_DELAY: Duration = Duration.ofMinutes(2L)
    const val FORK_ID_HASH_FIELD_NAME = "mfidh"
  }

  private val log: Logger = LogManager.getLogger(this::javaClass)

  private val privateKey = SecretKeyParser.fromLibP2pPrivKey(Bytes.wrap(privateKeyBytes))

  private val bootnodes =
    p2pConfig.discovery!!
      .bootnodes
      .map { NodeRecordFactory.DEFAULT.fromEnr(it) }
      .toList()

  private val discoverySystem: DiscoverySystem =
    DiscoverySystemBuilder()
      .listen(p2pConfig.ipAddress, p2pConfig.discovery!!.port.toInt())
      .secretKey(privateKey)
      .localNodeRecord(localNodeRecord())
      .bootnodes(bootnodes)
      .build()

  private val delayedExecutor: AsyncRunner =
    ScheduledExecutorAsyncRunner.create(
      "DiscoveryService",
      1,
      1,
      5,
      MetricTrackingExecutorFactory(metricsSystem),
    )

  private lateinit var bootnodeRefreshTask: Cancellable

  fun getLocalNodeRecord(): NodeRecord = discoverySystem.getLocalNodeRecord()

  init {
    val discoveryNetworkBuilder = DiscoverySystemBuilder()

    discoveryNetworkBuilder.listen(p2pConfig.ipAddress, p2pConfig.discovery!!.port.toInt())
    discoveryNetworkBuilder.secretKey(privateKey)
    discoveryNetworkBuilder.localNodeRecord(localNodeRecord())
    discoveryNetworkBuilder.bootnodes(bootnodes)
  }

  fun start() {
    discoverySystem
      .start()
      .thenRun {
        this.bootnodeRefreshTask =
          delayedExecutor.runWithFixedDelay(
            { this.pingBootnodes(bootnodes) },
            BOOTNODE_REFRESH_DELAY,
            { error: Throwable ->
              log.error(
                "Failed to contact discovery bootnodes",
                error,
              )
            },
          )
      }.get(30, TimeUnit.SECONDS)
    return
  }

  fun stop() {
    bootnodeRefreshTask.cancel()
    discoverySystem.stop()
  }

  fun updateForkIdHash(forkIdHash: Bytes) { // TODO: Need to call this when the fork id changes
    discoverySystem.updateCustomFieldValue(
      FORK_ID_HASH_FIELD_NAME,
      forkIdHash,
    )
  }

  fun searchForPeers(): SafeFuture<Collection<MaruDiscoveryPeer>> =
    discoverySystem
      .searchForNewPeers()
      // The current version of discovery doesn't return the found peers but next version will
      .toSafeFuture()
      .thenApply { getKnownPeers() }

  fun getKnownPeers(): Collection<MaruDiscoveryPeer> =
    discoverySystem
      .streamLiveNodes()
      .filter(this::checkNodeRecord)
      .map { node: NodeRecord ->
        convertSafeNodeRecordToDiscoveryPeer(node)
      }.toList()

  private fun convertSafeNodeRecordToDiscoveryPeer(node: NodeRecord): MaruDiscoveryPeer {
    // node record has been checked in checkNodeRecord, so we can convert to MaruDiscoveryPeer safely
    return MaruDiscoveryPeer(
      (node.get(EnrField.PKEY_SECP256K1) as Bytes),
      node.nodeId,
      node.tcpAddress.get(),
      node.get(FORK_ID_HASH_FIELD_NAME) as Bytes,
    )
  }

  private fun checkNodeRecord(node: NodeRecord): Boolean {
    if (node.get(FORK_ID_HASH_FIELD_NAME) == null) {
      log.info("Node record is missing forkId field: {}", node)
      return false
    }
    val forkId =
      (node.get(FORK_ID_HASH_FIELD_NAME) as? Bytes) ?: run {
        log.info("Failed to cast value for the forkId hash to Bytes")
        return false
      }
    if (forkId != Bytes.wrap(forkIdHashProvider.currentForkIdHash())) {
      log.info(
        "Peer {} is on a different chain. Expected: {}, Found: {}",
        node.nodeId,
        Bytes.wrap(forkIdHashProvider.currentForkIdHash()),
        forkId,
      )
      return false
    }
    if (node.get(EnrField.PKEY_SECP256K1) == null) {
      log.info("Node record is missing public key field: {}", node)
      return false
    }
    (node.get(EnrField.PKEY_SECP256K1) as? Bytes) ?: run {
      log.info("Failed to cast value for the public key to Bytes")
      return false
    }
    if (node.tcpAddress.isEmpty) {
      log.info(
        "node record doesn't have a TCP address: {}",
        node,
      )
      return false
    }
    return true
  }

  private fun pingBootnodes(bootnodeRecords: List<NodeRecord>) {
    bootnodeRecords.forEach(
      Consumer { bootnode: NodeRecord? ->
        SafeFuture
          .of(discoverySystem.ping(bootnode))
          .whenComplete { _, e ->
            if (e != null) {
              log.warn("Bootnode {} is unresponsive", bootnode)
            }
          }
      },
    )
  }

  private fun localNodeRecord(): NodeRecord {
    val nodeRecordBuilder: NodeRecordBuilder =
      NodeRecordBuilder()
        .secretKey(privateKey)
        .seq(UInt64.ONE)
        .address(
          p2pConfig.ipAddress,
          p2pConfig.discovery!!.port.toInt(),
          p2pConfig.port.toInt(),
        ).customField(FORK_ID_HASH_FIELD_NAME, Bytes.wrap(forkIdHashProvider.currentForkIdHash()))
    // TODO: do we want more custom fields to identify version/topics/role/something else?

    return nodeRecordBuilder.build()
  }
}
