/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import maru.core.SealedBeaconBlock
import maru.executionlayer.manager.ExecutionPayloadStatus
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import tech.pegasys.teku.infrastructure.async.SafeFuture

const val LINEA_DOMAIN = "linea"

// Mimicking ValidationResultCode for P2P communication
enum class ValidationResultCode {
  ACCEPT,
  REJECT,
  IGNORE,
}

sealed interface ValidationResult {
  val code: ValidationResultCode

  companion object {
    object Valid : ValidationResult {
      override val code = ValidationResultCode.ACCEPT
    }

    data class Invalid(
      val error: String,
      val cause: Throwable? = null,
    ) : ValidationResult {
      override val code = ValidationResultCode.REJECT
    }

    data class Ignore(
      val comment: String,
    ) : ValidationResult {
      override val code = ValidationResultCode.IGNORE
    }

    fun fromForkChoiceUpdatedResult(forkChoiceUpdatedResult: ForkChoiceUpdatedResult): ValidationResult {
      val payloadStatus = forkChoiceUpdatedResult.payloadStatus
      return when (payloadStatus.status) {
        ExecutionPayloadStatus.VALID -> Valid
        ExecutionPayloadStatus.INVALID -> Invalid(payloadStatus.validationError!!)
        else -> Ignore("Payload status is ${payloadStatus.status}")
      }
    }
  }
}

fun interface SealedBeaconBlockHandler<T> {
  fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<T>
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

  fun broadcastMessage(message: Message<*, GossipMessageType>): SafeFuture<*>

  /**
   * @return subscription id
   */
  fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler<ValidationResult>): Int

  fun unsubscribeFromBlocks(subscriptionId: Int)

  val port: UInt

  val nodeId: String

  val nodeAddresses: List<String>

  val discoveryAddresses: List<String>

  val enr: String?
}
