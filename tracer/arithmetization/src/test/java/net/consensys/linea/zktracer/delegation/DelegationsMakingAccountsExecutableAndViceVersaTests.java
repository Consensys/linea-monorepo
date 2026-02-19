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

import java.util.List;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.Test;

/**
 * These tests address issue <a
 * href="https://github.com/Consensys/linea-monorepo/issues/2355">[ZkTracer] Test transactions where
 * delegations swap the recipient from executable to non-executable and vice versa</a>
 */
public class DelegationsMakingAccountsExecutableAndViceVersaTests extends TracerTestBase {

  static List<ToyAccount> accounts;

  static final ToyAccount smcWithGenericCode =
      ToyAccount.builder()
          .address(Address.fromHexString("0xc0dec0ffee"))
          .nonce(0)
          .balance(Wei.fromEth(1))
          .code(Bytes.fromHexString("0x600160005260206000f3"))
          .build();

  static final ToyAccount smcWithDelegationPrefixButWrongSize =
      ToyAccount.builder()
          .address(Address.fromHexString("0xdeadc0de"))
          .nonce(0)
          .balance(Wei.fromEth(1))
          .code(Bytes.fromHexString("0xef0100" + "5b".repeat(7)))
          .build();

  static final ToyAccount smcWithCodeSize23 =
      ToyAccount.builder()
          .address(Address.fromHexString("0xc0debeef"))
          .nonce(0)
          .balance(Wei.fromEth(1))
          .code(Bytes.fromHexString("0x6001" + "5b".repeat(21)))
          .build();

  static final ToyAccount smcWithCodeSize23AndCloseMissOnPrefix =
      ToyAccount.builder()
          .address(Address.fromHexString("0xdeadbeefc0de"))
          .nonce(0)
          .balance(Wei.fromEth(1))
          .code(Bytes.fromHexString("0xef01ff" + "5b".repeat(20)))
          .build();

  /**
   * The transaction recipient starts out having <b>nonempty byte code</b>, <b>delegation byte
   * code</b> to be precise. The target of the delegation being a proper smart contract i.e. an
   * account with
   *
   * <ul>
   *   <li>nonempty byte code with either
   *       <ul>
   *         <li>has size <b>≠ 23</b> or
   *         <li>has size <b>= 23</b> but doesn't start with <b>0xef0100</b>
   *       </ul>
   * </ul>
   *
   * A delegation renders the recipient of the transaction non-executable by delegating it to a
   * proper (yet uninteresting) smart contract.
   */
  @Test
  void delegationToEoaMakesAccountNonExecutableTest() {}

  @Test
  void delegationToPrcMakesAccountNonExecutableTest() {}

  @Test
  void delegationResetMakesAccountNonExecutableTest() {}

  /**
   * The transaction recipient starts out having <b>empty byte code</b>. A delegation renders the
   * recipient of the transaction executable by delegating it to a proper (yet uninteresting) smart
   * contract.
   */
  @Test
  void delegationToSmcMakesAccountExecutableTest() {}

  /**
   * The transaction recipient starts out having <b>empty byte code</b>. A delegation renders the
   * recipient of the transaction executable by delegating it to delegated EOA. This delegation
   * makes the target code "executable".
   *
   * <p><b>Note.</b> The target code that will be run is itself "delegation code" and thus starts
   * with <b>0xef</b>. Running such "delegation code" immediately raises an
   * <b>invalidOpCodeException</b> which stops execution with an exception.
   */
  @Test
  void delegationToDelegatedEoaMakesAccountExecutableTest() {}

  void populateAccounts() {
    accounts.add(smcWithGenericCode);
    accounts.add(smcWithDelegationPrefixButWrongSize);
    accounts.add(smcWithCodeSize23);
    accounts.add(smcWithCodeSize23AndCloseMissOnPrefix);
    accounts.add(senderAccount);
    accounts.add(authorityAccount);
  }
}
