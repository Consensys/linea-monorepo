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
package net.consensys.linea.zktracer.delegation;

import static net.consensys.linea.zktracer.delegation.Utils.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

/**
 * These tests address issue <a
 * href="https://github.com/Consensys/linea-monorepo/issues/2437">[ZkTracer] Test delegation lists +
 * access lists</a>
 */
public class DelegationAndAccessListTests extends TracerTestBase {

  @ParameterizedTest
  @MethodSource("delegationAndAccessListScenarios")
  void delegationsAndAccessListTests(
      ChainIdValidity chainIdValidity,
      AuthorityInAccessList authorityInAccessList,
      RequiresEvmExecution requiresEvmExecution,
      AuthorityExistence authorityExistence,
      TouchAuthority touchAuthority,
      TouchingMethod touchingMethod,
      TestInfo testInfo) {

    runTestWithParameters(
        chainIdValidity,
        authorityInAccessList,
        requiresEvmExecution,
        authorityExistence,
        touchAuthority,
        touchingMethod,
        testInfo);
  }

  @Test
  void targetedTest(TestInfo testInfo) {

    runTestWithParameters(
        ChainIdValidity.DELEGATION_CHAIN_ID_IS_NETWORK_CHAIN_ID,
        AuthorityInAccessList.AUTHORITY_NOT_IN_ACCESS_LIST,
        RequiresEvmExecution.REQUIRES_EVM_EXECUTION,
        AuthorityExistence.AUTHORITY_DOES_NOT_EXIST,
        TouchAuthority.EXECUTION_TOUCHES_AUTHORITY,
        TouchingMethod.EXTCODESIZE,
        testInfo);
  }

  void runTestWithParameters(
      ChainIdValidity chainIdValidity,
      AuthorityInAccessList authorityInAccessList,
      RequiresEvmExecution requiresEvmExecution,
      AuthorityExistence authorityExistence,
      TouchAuthority touchAuthority,
      TouchingMethod touchingMethod,
      TestInfo testInfo) {

    if (authorityInAccessList == AuthorityInAccessList.AUTHORITY_IN_ACCESS_LIST) {
      List<AccessListEntry> accessList = new ArrayList<>();
      accessList.add(AccessListEntry.createAccessListEntry(authorityAddress, new ArrayList<>()));
      tx.accessList(accessList);
    }

    tx.clearCodeDelegations();
    tx.addCodeDelegation(
        chainIdValidity.tupleChainId(),
        delegationAddress,
        authorityExistence.tupleNonce(),
        authorityKeyPair);

    smcAccount.setCode(
        codeThatMayTouchAuthority(requiresEvmExecution, touchAuthority, touchingMethod).compile());

    final List<ToyAccount> accounts = new ArrayList<>();
    accounts.add(senderAccount);
    accounts.add(smcAccount);
    if (authorityExistence == Utils.AuthorityExistence.AUTHORITY_EXISTS) {
      accounts.add(authorityAccount);
    }

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(accounts)
        .transaction(tx.build())
        .zkTracerValidator(zkTracer -> {})
        .build()
        .run();
  }

  private static Stream<Arguments> delegationAndAccessListScenarios() {
    List<Arguments> argumentsList = new ArrayList<>();

    for (ChainIdValidity chainIdValidity : ChainIdValidity.values()) {
      for (AuthorityInAccessList authorityInAccessList : AuthorityInAccessList.values()) {
        for (RequiresEvmExecution transactionRequiresEvmExecution : RequiresEvmExecution.values()) {
          for (AuthorityExistence authorityExistence : AuthorityExistence.values()) {
            for (TouchAuthority touchAuthority : TouchAuthority.values()) {
              for (TouchingMethod touchingMethod : TouchingMethod.values()) {
                argumentsList.add(
                    Arguments.of(
                        chainIdValidity,
                        authorityInAccessList,
                        transactionRequiresEvmExecution,
                        authorityExistence,
                        touchAuthority,
                        touchingMethod));
              }
            }
          }
        }
      }
    }
    return argumentsList.stream();
  }
}
