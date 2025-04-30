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
package maru.consensus.qbft.adapters

import java.time.Clock
import maru.consensus.qbft.toAddress
import maru.database.BeaconChain
import org.hyperledger.besu.consensus.common.bft.BftHelpers
import org.hyperledger.besu.consensus.common.bft.BlockTimer
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.RoundTimer
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCreatorFactory
import org.hyperledger.besu.consensus.qbft.core.types.QbftFinalState
import org.hyperledger.besu.consensus.qbft.core.types.QbftValidatorProvider
import org.hyperledger.besu.cryptoservices.NodeKey
import org.hyperledger.besu.datatypes.Address

class QbftFinalStateAdapter(
  private val localAddress: Address,
  private val nodeKey: NodeKey,
  private val validatorProvider: QbftValidatorProvider,
  private val proposerSelector: ProposerSelector,
  private val validatorMulticaster: ValidatorMulticaster,
  private val roundTimer: RoundTimer,
  private val blockTimer: BlockTimer,
  private val blockCreatorFactory: QbftBlockCreatorFactory,
  private val clock: Clock,
  private val beaconChain: BeaconChain,
) : QbftFinalState {
  override fun getValidatorMulticaster(): ValidatorMulticaster = validatorMulticaster

  override fun getNodeKey(): NodeKey = nodeKey

  override fun getRoundTimer(): RoundTimer = roundTimer

  override fun isLocalNodeValidator(): Boolean = getValidators().contains(localAddress)

  override fun getValidators(): Collection<Address> =
    beaconChain.getLatestBeaconState().validators.map { it.toAddress() }

  override fun getLocalAddress(): Address = localAddress

  override fun getClock(): Clock = clock

  override fun getBlockCreatorFactory(): QbftBlockCreatorFactory = blockCreatorFactory

  override fun getQuorum(): Int = BftHelpers.calculateRequiredValidatorQuorum(getValidators().size)

  override fun getBlockTimer(): BlockTimer = blockTimer

  override fun isLocalNodeProposerForRound(roundIdentifier: ConsensusRoundIdentifier): Boolean =
    proposerSelector.selectProposerForRound(roundIdentifier).equals(localAddress)
}
