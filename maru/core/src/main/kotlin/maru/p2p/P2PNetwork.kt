/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.p2p

import maru.core.SealedBeaconBlock
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface SealedBlockHandler {
  fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*>
}

/**
 * Interface for the P2P Network functionality.
 *
 * This network is used to:
 * 1. Exchange new blocks between nodes
 * 2. Sync new nodes joining the network
 * 3. Exchange QBFT messages between QBFT Validators
 */
interface P2PNetwork {
  /**
   * Start the P2P network service.
   *
   * @return A future that completes when the service is started.
   */
  fun start(): SafeFuture<Unit>

  /**
   * Stop the P2P network service.
   *
   * @return A future that completes when the service is stopped.
   */
  fun stop(): SafeFuture<Unit>

  fun broadcastMessage(message: Message<*>)

  fun subscribeToBlocks(subscriber: SealedBlockHandler)
}
