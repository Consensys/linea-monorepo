/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.statemanager;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.testing.MultiBlockExecutionEnvironment;
import net.consensys.linea.testing.TransactionProcessingResultValidator;
import net.consensys.linea.testing.generated.FrameworkEntrypoint;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

public class TransactionStorageTest {
  TestContext tc;

  @Test
  void testTransactionMapStorage() {
    // initialize the test context
    this.tc = new TestContext();
    this.tc.initializeTestContext();
    // prepare the transaction validator
    TransactionProcessingResultValidator resultValidator =
        new StateManagerTestValidator(
            tc.frameworkEntryPointAccount,
            // Creates, writes, reads and self-destructs generate 2 logs,
            // Reverted operations only have 1 log
            List.of(7, 7));
    // fetch the Hub metadata for the state manager maps
    // StateManagerMetadata stateManagerMetadata = Hub.stateManagerMetadata();

    // prepare a multi-block execution of transactions
    final MultiBlockExecutionEnvironment multiBlockEnv =
        MultiBlockExecutionEnvironment.builder()
            // initialize accounts
            .accounts(
                List.of(
                    tc.initialAccounts[0],
                    tc.externallyOwnedAccounts[0],
                    tc.initialAccounts[2],
                    tc.frameworkEntryPointAccount))
            // Block 1
            .addBlock(
                List.of(
                    tc.newTxFromCalls(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        new FrameworkEntrypoint.ContractCall[] {
                          tc.writeToStorageCall(tc.addresses[0], 3L, 1L, false, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 2L, false, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 3L, false, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 1234L, true, BigInteger.ONE),
                        }),
                    tc.newTxFromCalls(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        new FrameworkEntrypoint.ContractCall[] {
                          tc.writeToStorageCall(tc.addresses[0], 3L, 1234L, true, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 4L, false, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 5L, false, BigInteger.ONE),
                          tc.writeToStorageCall(tc.addresses[0], 3L, 6L, false, BigInteger.ONE),
                        })))
            .transactionProcessingResultValidator(resultValidator)
            .build();
    multiBlockEnv.run();

    // Initialize the storageFirstAndLastMap list
    List<Map<Map<Address, EWord>, FragmentFirstAndLast<StorageFragment>>>
        storageFirstAndLastMapList = new ArrayList<>();

    // We count the number of transactions in the hub
    int txCount = multiBlockEnv.getHub().state().txCount();
    // We iterate over the transactions
    for (int txNb = 0; txNb < txCount; txNb++) {
      // We create an storageFirstAndLastMap for each transaction
      storageFirstAndLastMapList.add(new HashMap<>());
      // We retrieve the trace section list
      List<TraceSection> traceSectionList =
          multiBlockEnv
              .getHub()
              .state()
              .getState()
              .operationsInTransactionBundle()
              .get(txNb)
              .traceSections()
              .trace();
      // For each trace section
      for (TraceSection traceSection : traceSectionList) {
        // We iterate over the fragments
        for (TraceFragment traceFragment : traceSection.fragments()) {
          // We cast them to StorageFragment
          // If an exception occurs, it means the Fragment is not a StorageFragment so we
          // disregard it and continue
          try {
            StorageFragment storageFragment = (StorageFragment) traceFragment;
            Address address = storageFragment.getStorageSlotIdentifier().getAddress();
            EWord key = storageFragment.getStorageSlotIdentifier().getStorageKey();
            // We update the storageFirstAndLastMapList
            updateStorageFirstAndLast(
                storageFragment, storageFirstAndLastMapList.get(txNb), Map.of(address, key));
          } catch (Exception e) {
            // ignore
          }
        }
      }
    }

    // prepare data for asserts
    // expected first values for the keys we are testing
    int noBlocks = 3;
    EWord[][] expectedFirst = {
      {
        EWord.of(0L),
      },
      {
        EWord.of(3L),
      },
    };
    // expected last values for the keys we are testing
    EWord[][] expectedLast = {
      {
        EWord.of(3L),
      },
      {
        EWord.of(6L),
      },
    };
    // prepare the key pairs
    List<Map<Address, EWord>> addrStorageKeyMap =
        List.of(Map.of(tc.initialAccounts[0].getAddress(), EWord.of(3L)));

    for (int txCounter = 0; txCounter < txCount; txCounter++) {
      Map<Map<Address, EWord>, FragmentFirstAndLast<StorageFragment>> storageMap =
          storageFirstAndLastMapList.get(txCounter);
      for (int i = 0; i < addrStorageKeyMap.size(); i++) {
        FragmentFirstAndLast<StorageFragment> storageData =
            storageMap.get(addrStorageKeyMap.get(i));
        // asserts for the first and last storage values in conflation
        // -1 due to block numbering
        assertEquals(expectedFirst[txCounter][i], storageData.getFirst().getValueCurrent());
        assertEquals(expectedLast[txCounter][i], storageData.getLast().getValueNext());
      }
    }

    System.out.println("Done");
  }

  public void updateStorageFirstAndLast(
      StorageFragment fragment,
      Map<Map<Address, EWord>, FragmentFirstAndLast<StorageFragment>> storageFirstAndLastMap,
      Map<Address, EWord> key) {
    // Setting the post transaction first and last value
    int dom = fragment.getDomSubStampsSubFragment().domStamp();
    int sub = fragment.getDomSubStampsSubFragment().subStamp();

    if (!storageFirstAndLastMap.containsKey(key)) {
      FragmentFirstAndLast<StorageFragment> txnFirstAndLast =
          new FragmentFirstAndLast<StorageFragment>(fragment, fragment, dom, sub, dom, sub);
      storageFirstAndLastMap.put(key, txnFirstAndLast);
    } else {
      // the storage key has already been accessed for this account
      FragmentFirstAndLast<StorageFragment> txnFirstAndLast = storageFirstAndLastMap.get(key);
      // Replace condition
      if (FragmentFirstAndLast.strictlySmallerStamps(
          txnFirstAndLast.getLastDom(), txnFirstAndLast.getLastSub(), dom, sub)) {
        txnFirstAndLast.setLast(fragment);
        txnFirstAndLast.setLastDom(dom);
        txnFirstAndLast.setLastSub(sub);
        storageFirstAndLastMap.put(key, txnFirstAndLast);
      }
    }
  }
}
