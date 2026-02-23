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

package net.consensys.linea.zktracer.instructionprocessing.callTests;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class CallDelegationTests extends TracerTestBase {

  static final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  static final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(42)
          .address(senderAddress)
          .keyPair(senderKeyPair)
          .build();

  static final ToyAccount rootAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(67)
          .address(Address.fromHexString("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
          .build();

  static final ToyAccount callerAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(3))
          .nonce(69)
          .address(Address.fromHexString("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"))
          .build();

  static final ToyAccount calleeAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(90)
          .address(Address.fromHexString("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"))
          .build();

  static final ToyAccount smcAccount1 =
      ToyAccount.builder()
          .balance(Wei.fromEth(5))
          .nonce(101)
          .address(Address.fromHexString("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"))
          .build();

  static final ToyAccount smcAccount2 =
      ToyAccount.builder()
          .balance(Wei.fromEth(6))
          .nonce(666)
          .address(Address.fromHexString("0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"))
          /*
          .storage(
              new HashMap<>() {
                {
                  put(UInt256.valueOf(0), UInt256.valueOf(0));
                }
              }) for testing purposes we may also start with a non-empty storage */
          .build();

  BytecodeCompiler conditionalCallProgram =
      BytecodeCompiler.newProgram(chainConfig)
          .push(0)
          .op(OpCode.SLOAD) // LOOP_DEPTH_CURRENT
          .push(3) // LOOP_DEPTH_MAX
          .op(OpCode.GT) // LOOP_DEPTH_MAX > LOOP_DEPTH_CURRENT
          .push(10)
          .op(OpCode.JUMPI) // if LOOP_DEPTH_CURRENT < LOOP_DEPTH_MAX jump to JUMPDEST else STOP
          .op(OpCode.STOP)
          .op(OpCode.JUMPDEST) // PC = 10
          .push(0)
          .op(OpCode.SLOAD)
          .push(1)
          .op(OpCode.ADD)
          .push(0)
          .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
          // execute the call
          .push(0) // return at capacity
          .push("ff".repeat(32)) // return at offset
          .push(0) // call data size
          .push(0) // call data offset
          .push("ca11ee") // address
          .push(1000) // gas
          .op(OpCode.STATICCALL);

  // apply to either the root or the caller
  public enum RevertType {
    TERMINATES_ON_REVERT,
    TERMINATES_ON_NON_REVERT;
  }

  // apply to either the root or the caller
  public enum LoopType {
    INFINITE_LOOP,
    EXIT_EARLY;
  }

  // root -> DELEGATED  ==  root -> [EOA that is delegated to a SMC running the caller code]
  // root -> SMC        ==  root -> [SMC containing the caller code]
  public enum CallerType {
    DELEGATED,
    SMC, // this exists, today
    ;
  }

  public enum CalleeType {
    // the first few we don't really care about: they don't lead to execution
    DELEGATED_TO_NON_EXISTENT,
    DELEGATED_TO_EMPTY_CODE_ACCOUNT,
    DELEGATED_TO_PRC,
    DELEGATED_TO_SELF,
    // the ones below are the ones we really care about, they introduce circularity etc ...
    DELEGATED_TO_ROOT,
    DELEGATED_TO_CALLER,
    DELEGATED_TO_SMC,
    SMC, // this exists, today
    ;
  }

  @Test
  public void testCallDelegation() {}
}
