/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.pipeline

import java.util.function.Consumer
import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.p2p.ValidationResult
import maru.p2p.ValidationResultCode
import maru.p2p.ValidationResultCode.ACCEPT
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason

class ImportBlocksStep(
  private val blockImporter: SealedBeaconBlockImporter<ValidationResult>,
) : Consumer<List<SealedBlockWithPeer>> {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  override fun accept(blocksWithPeers: List<SealedBlockWithPeer>) {
    // Process blocks sequentially
    blocksWithPeers.forEach { blockAndPeer ->
      try {
        val result = blockImporter.importBlock(blockAndPeer.sealedBeaconBlock).join()
        when (result.code) {
          ACCEPT -> {
            log.info(
              "Successfully imported block number={} hash={}",
              blockAndPeer.sealedBeaconBlock.beaconBlock.beaconBlockHeader.number,
              blockAndPeer.sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash,
            )
          }
          ValidationResultCode.REJECT -> {
            blockAndPeer.peer.disconnectCleanly(DisconnectReason.REMOTE_FAULT)
            log.error(
              "Block validation failed for block {}",
              blockAndPeer.sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash,
            )
            return
          }
          ValidationResultCode.IGNORE -> {
            log.warn(
              "Block validation ignored for block {}",
              blockAndPeer.sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash,
            )
            return
          }
        }
      } catch (e: Exception) {
        log.error("Exception importing block", e)
        throw e
      }
    }
  }
}
