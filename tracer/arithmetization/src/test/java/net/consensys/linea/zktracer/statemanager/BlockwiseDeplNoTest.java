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
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.testing.MultiBlockExecutionEnvironment;
import net.consensys.linea.testing.TransactionProcessingResultValidator;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

public class BlockwiseDeplNoTest {
  TestContext tc;

  @Test
  void testBlockwiseDeplNo() {
    // initialize the test context
    this.tc = new TestContext();
    this.tc.initializeTestContext();
    // prepare the transaction validator
    TransactionProcessingResultValidator resultValidator =
        new StateManagerTestValidator(
            tc.frameworkEntryPointAccount,
            // Creates, writes, reads and self-destructs generate 2 logs,
            // Reverted operations only have 1 log
            List.of(2, 2, 2, 2, 2, 2, 2, 2, 1, 2, 1));
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
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[0],
                        TestContext.snippetsCodeForCreate2,
                        false),
                    tc.selfDestruct(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.newAddresses[0],
                        tc.frameworkEntryPointAddress,
                        false,
                        BigInteger.ONE),
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[0],
                        TestContext.snippetsCodeForCreate2,
                        false),
                    tc.selfDestruct(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.newAddresses[0],
                        tc.frameworkEntryPointAddress,
                        false,
                        BigInteger.ONE)))
            .addBlock(
                List.of(
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[1],
                        TestContext.snippetsCodeForCreate2,
                        false),
                    tc.selfDestruct(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.newAddresses[1],
                        tc.frameworkEntryPointAddress,
                        false,
                        BigInteger.ONE)))
            .addBlock(
                List.of( // test some reverted calls
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[2],
                        TestContext.snippetsCodeForCreate2,
                        false),
                    tc.selfDestruct(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.newAddresses[2],
                        tc.frameworkEntryPointAddress,
                        false,
                        BigInteger.ONE),
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[2],
                        TestContext.snippetsCodeForCreate2,
                        true),
                    tc.deployWithCreate2_withRevertTrigger(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.frameworkEntryPointAddress,
                        tc.salts[2],
                        TestContext.snippetsCodeForCreate2,
                        false),
                    tc.selfDestruct(
                        tc.externallyOwnedAccounts[0],
                        tc.keyPairs[0],
                        tc.newAddresses[2],
                        tc.frameworkEntryPointAddress,
                        true,
                        BigInteger.ONE)
                    // since the last self-destruct gets reverted, the last call will not increase
                    // the
                    // deplNo
                    ))
            .transactionProcessingResultValidator(resultValidator)
            .build();

    multiBlockEnv.run();

    // As the state manager feature is not available anymore, we are not able to keep track of the
    // deployment number per address
    // However, in the trace we have all the sections and fragments to replay and record the values
    // that the state manager used to keep track of while tracing was ongoing

    // Map to keep track of the deployment number per address
    Map<Address, Map<Integer, Integer>> minDeplNoBlock = new HashMap<>();
    Map<Address, Map<Integer, Integer>> maxDeplNoBlock = new HashMap<>();

    // We count the number of transactions in the hub, here we have 11
    int txCount = multiBlockEnv.getHub().state().txCount();
    for (int txNb = 0; txNb < txCount; txNb++) {
      // Relative block number is constant per transaction
      // Tx number starts from 1
      int relBlokNo =
          multiBlockEnv
              .getHub()
              .txStack()
              .getByAbsoluteTransactionNumber(txNb + 1)
              .getRelativeBlockNumber();
      // We retrieve the trace section list for each transaction
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
          // We cast them to AccountFragment
          // If an exception occurs, it means the Fragment is not an AccountFragment so we
          // disregard it and continue
          try {
            AccountFragment accountFragment = (AccountFragment) traceFragment;
            Address address = accountFragment.oldState().address();
            int deplNo = accountFragment.newState().deploymentNumber();
            updateDeplNoBlockMaps(address, relBlokNo, deplNo, minDeplNoBlock, maxDeplNoBlock);
          } catch (Exception e) {
            // ignore
          }
        }
      }
    }

    // prepare data for asserts
    // expected first values for the keys we are testing
    int noBlocks = 3;
    Integer[][] expectedMin = {
      {1, null, null},
      {null, 1, null},
      {null, null, 1},
    };
    // expected last values for the keys we are testing
    Integer[][] expectedMax = {
      {
        4, null, null,
      },
      {
        null, 2, null,
      },
      {null, null, 4},
    };
    // prepare the key pairs
    Address[] keys = {
      tc.newAddresses[0], tc.newAddresses[1], tc.newAddresses[2],
    };

    // blocks are numbered starting from 1
    for (int block = 1; block <= noBlocks; block++) {
      for (int i = 0; i < keys.length; i++) {
        Integer minNo = minDeplNoBlock.get(keys[i]).get(block);
        Integer maxNo = maxDeplNoBlock.get(keys[i]).get(block);
        // asserts for the first and last storage values in conflation
        // -1 due to block numbering
        assertEquals(expectedMin[block - 1][i], minNo);
        assertEquals(expectedMax[block - 1][i], maxNo);
      }
    }

    System.out.println("Done");
  }

  public void updateDeplNoBlockMaps(
      Address address,
      int blockNumber,
      int currentDeplNo,
      Map<Address, Map<Integer, Integer>> minDeplNoBlock,
      Map<Address, Map<Integer, Integer>> maxDeplNoBlock) {
    if (minDeplNoBlock.containsKey(address)
        && minDeplNoBlock.get(address).containsKey(blockNumber)) {
      // the maps already contain deployment info for this address, and this is not the first one in
      // the block
      // since it is not the first, we do not update the minDeplNoBlock
      // but we update the maxDeplNoBlock
      maxDeplNoBlock.put(address, new HashMap<>());
      maxDeplNoBlock.get(address).put(blockNumber, currentDeplNo);
    } else {
      // this is the first time we have a deployment at this address in the block
      minDeplNoBlock.put(address, new HashMap<>());
      minDeplNoBlock.get(address).put(blockNumber, currentDeplNo);
      maxDeplNoBlock.put(address, new HashMap<>());
      maxDeplNoBlock.get(address).put(blockNumber, currentDeplNo);
    }
  }
}
