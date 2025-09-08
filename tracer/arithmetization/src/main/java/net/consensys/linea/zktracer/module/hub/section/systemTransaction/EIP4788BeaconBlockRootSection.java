/*
 * Copyright ConsenSys Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.zktracer.module.hub.section.systemTransaction;

import static net.consensys.linea.zktracer.Trace.*;
import static net.consensys.linea.zktracer.module.hub.TransactionProcessingType.SYSI;
import static net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment.systemTransactionStoring;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes16;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment;
import net.consensys.linea.zktracer.module.hub.fragment.transaction.system.Eip4788TransactionFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.types.AddressUtils;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

public class EIP4788BeaconBlockRootSection extends TraceSection {

  public static final Address EIP4788_BEACONROOT_ADDRESS =
      AddressUtils.addressFromBytes(
          Bytes.concatenate(
              Bytes.minimalBytes(BEACON_ROOTS_ADDRESS_HI),
              bigIntegerToBytes16(BEACON_ROOTS_ADDRESS_LO)));

  public EIP4788BeaconBlockRootSection(
      Hub hub, WorldView world, ProcessableBlockHeader blockHeader) {
    super(hub, (short) 5);
    final AccountSnapshot beaconrootAccount =
        AccountSnapshot.canonical(hub, world, EIP4788_BEACONROOT_ADDRESS, false);
    final long timestamp = blockHeader.getTimestamp();
    final boolean currentBlockIsGenesisBlock = blockHeader.getNumber() == 0;
    final boolean isNonTrivialOperation =
        !currentBlockIsGenesisBlock && !beaconrootAccount.code().isEmpty();
    final Bytes32 beaconRoot =
        isNonTrivialOperation ? blockHeader.getParentBeaconBlockRoot().get() : Bytes32.ZERO;

    final Eip4788TransactionFragment transactionFragment =
        new Eip4788TransactionFragment(timestamp, beaconRoot, currentBlockIsGenesisBlock);
    fragments().add(transactionFragment);
    hub.txnData().callTxnDataForSystemTransaction(transactionFragment.type());

    final AccountFragment accountFragment =
        hub.factories()
            .accountFragment()
            .makeWithTrm(
                beaconrootAccount,
                beaconrootAccount,
                EIP4788_BEACONROOT_ADDRESS,
                DomSubStampsSubFragment.standardDomSubStamps(hubStamp(), 1),
                SYSI);
    fragments().add(accountFragment);

    if (isNonTrivialOperation) {
      final EWord keyTimestamp = EWord.of(timestamp % HISTORY_BUFFER_LENGTH);
      final StorageFragment storingTimestamp =
          systemTransactionStoring(
              hub,
              EIP4788_BEACONROOT_ADDRESS,
              keyTimestamp,
              EWord.of(
                  world
                      .get(EIP4788_BEACONROOT_ADDRESS)
                      .getStorageValue(UInt256.fromBytes(keyTimestamp))),
              EWord.of(timestamp),
              2);
      fragments().add(storingTimestamp);

      final EWord keyBeaconRoot =
          EWord.of((timestamp % HISTORY_BUFFER_LENGTH) + HISTORY_BUFFER_LENGTH);
      final StorageFragment storingBeaconroot =
          systemTransactionStoring(
              hub,
              EIP4788_BEACONROOT_ADDRESS,
              keyBeaconRoot,
              EWord.of(
                  world
                      .get(EIP4788_BEACONROOT_ADDRESS)
                      .getStorageValue(UInt256.fromBytes(keyBeaconRoot))),
              EWord.of(beaconRoot),
              3);
      fragments().add(storingBeaconroot);
    }

    fragments().add(ContextFragment.readZeroContextData(hub));
  }
}
