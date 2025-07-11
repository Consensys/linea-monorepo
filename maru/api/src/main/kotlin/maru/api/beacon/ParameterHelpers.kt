/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.api.beacon

import maru.api.BlockNotFoundException
import maru.api.ChainDataProvider
import maru.api.HandlerException
import maru.core.SealedBeaconBlock
import maru.extensions.isPositiveNumber

fun getBlock(
  blockId: String,
  chainDataProvider: ChainDataProvider,
): SealedBeaconBlock {
  val maruSealedBeaconBlock =
    try {
      when {
        blockId.lowercase() == "head" -> chainDataProvider.getLatestBeaconBlock()
        blockId.lowercase() == "finalized" -> chainDataProvider.getLatestBeaconBlock()
        blockId.lowercase() == "genesis" -> chainDataProvider.getGenesisBeaconBlock()
        blockId.isPositiveNumber() -> chainDataProvider.getBeaconBlockByNumber(blockId.toLong().toULong())
        blockId.startsWith("0x") -> chainDataProvider.getBeaconBlockByBlockRoot(blockId)
        else -> throw HandlerException(400, "Invalid block ID: $blockId")
      }
    } catch (e: BlockNotFoundException) {
      throw HandlerException(404, "Block not found")
    }
  return maruSealedBeaconBlock
}
