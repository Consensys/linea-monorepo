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

import maru.consensus.validation.BeaconBlockValidatorFactory
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockImporter
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockValidator
import org.hyperledger.besu.consensus.qbft.core.types.QbftProtocolSchedule

class QbftProtocolScheduleAdapter(
  private val blockImporter: QbftBlockImporter,
  private val beaconBlockValidatorFactory: BeaconBlockValidatorFactory,
) : QbftProtocolSchedule {
  override fun getBlockImporter(blockHeader: QbftBlockHeader): QbftBlockImporter = blockImporter

  override fun getBlockValidator(blockHeader: QbftBlockHeader): QbftBlockValidator {
    val beaconBlockHeader = blockHeader.toBeaconBlockHeader()
    return QbftBlockValidatorAdapter(beaconBlockValidatorFactory.createValidatorForBlock(beaconBlockHeader))
  }
}
