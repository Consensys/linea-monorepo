/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import java.util.Optional
import maru.consensus.validation.BlockValidator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockValidator

class QbftBlockValidatorAdapter(
  private val blockValidator: BlockValidator,
) : QbftBlockValidator {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  override fun validateBlock(qbftBlock: QbftBlock): QbftBlockValidator.ValidationResult {
    log.trace("validating ${blockValidator.javaClass.canonicalName}")
    val beaconBlock = qbftBlock.toBeaconBlock()
    return when (val blockValidationResult = blockValidator.validateBlock(beaconBlock).get()) {
      is Ok -> QbftBlockValidator.ValidationResult(true, Optional.empty())
      is Err ->
        QbftBlockValidator.ValidationResult(
          false,
          Optional.of(blockValidationResult.error.toString()),
        )
    }
  }
}
