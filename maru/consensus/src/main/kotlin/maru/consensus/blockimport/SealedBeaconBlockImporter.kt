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
package maru.consensus.blockimport

import maru.consensus.state.StateTransition
import maru.core.SealedBeaconBlock
import maru.database.BeaconChain
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface SealedBeaconBlockImporter {
  fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*>
}

/**
 * Responsible for: transactional  and El node
 * 1. state transition of node's BeaconChain
 * 2. new block import into an EL node
 * The import is transactional, I.e. all or nothing approach
 */
class TransactionalSealedBeaconBlockImporter(
  private val beaconChain: BeaconChain,
  private val stateTransition: StateTransition,
  private val beaconBlockImporter: BeaconBlockImporter,
) : SealedBeaconBlockImporter {
  private val log: Logger = LogManager.getLogger(this::javaClass)

  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<ForkChoiceUpdatedResult> {
    val updater = beaconChain.newUpdater()
    try {
      return stateTransition
        .processBlock(sealedBeaconBlock.beaconBlock)
        .thenCompose { resultingState ->
          updater
            .putBeaconState(resultingState)
            .putSealedBeaconBlock(sealedBeaconBlock)
          beaconBlockImporter
            .importBlock(resultingState, sealedBeaconBlock.beaconBlock)
        }.thenApply {
          updater.commit()
          it
        }.whenException {
          updater.rollback()
        }.whenComplete { _, _ ->
          updater.close()
        }
    } catch (e: Exception) {
      log.error("Block import state transition failed!: ${e.message}", e)
      return SafeFuture.failedFuture(e)
    }
  }
}

/**
 * Verifies the seal and delegates to another beaconBlockImporter
 */
class VerifyingSealedBeaconBlockImporter(
  private val beaconBlockImporter: SealedBeaconBlockImporter,
) : SealedBeaconBlockImporter {
  // TODO: implement seal verification
  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*> =
    beaconBlockImporter.importBlock(sealedBeaconBlock)
}
