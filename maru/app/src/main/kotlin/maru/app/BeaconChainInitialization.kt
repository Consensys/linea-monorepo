/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.EMPTY_HASH
import maru.core.HashUtil
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.database.BeaconChain
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.RLPSerializers
import maru.serialization.rlp.stateRoot
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter

class BeaconChainInitialization(
  private val executionLayerClient: Web3j,
  private val beaconChain: BeaconChain,
) {
  private fun initializeDb(validatorSet: Set<Validator>) {
    val genesisExecutionPayload =
      executionLayerClient
        .ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), true)
        .send()
        .block
        .toDomain()

    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptySet(), executionPayload = genesisExecutionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = genesisExecutionPayload.timestamp,
        proposer = Validator(genesisExecutionPayload.feeRecipient),
        parentRoot = EMPTY_HASH,
        stateRoot = EMPTY_HASH,
        bodyRoot = EMPTY_HASH,
        headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
      )

    val tmpGenesisStateRoot =
      BeaconState(
        latestBeaconBlockHeader = beaconBlockHeader,
        validators = validatorSet,
      )
    val stateRootHash = HashUtil.stateRoot(tmpGenesisStateRoot)

    val genesisBlockHeader = beaconBlockHeader.copy(stateRoot = stateRootHash)
    val genesisBlock = BeaconBlock(genesisBlockHeader, beaconBlockBody)
    val genesisState = BeaconState(genesisBlockHeader, validatorSet)
    beaconChain.newUpdater().run {
      putBeaconState(genesisState)
      putSealedBeaconBlock(SealedBeaconBlock(genesisBlock, emptySet()))
      commit()
    }
  }

  fun ensureDbIsInitialized(validatorSet: Set<Validator>) {
    if (!beaconChain.isInitialized()) {
      initializeDb(validatorSet)
    }
  }
}
