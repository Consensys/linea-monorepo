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

package net.consensys.linea.zktracer.forkSpecific.systemTransaction;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;

public class SystemTransactionTestUtils {

  static Bytes byteCodeCallingBeaconRootSystemAccount(
      ChainConfig chainConfig, Address systemContractAddress, long arg) {
    return BytecodeCompiler.newProgram(chainConfig)
        // prepare memory with arg left padded
        .push(Bytes32.leftPad(Bytes.minimalBytes(arg))) // value
        .push(0) // offset
        .op(OpCode.MSTORE)
        // call system contract
        .push(0) // retSize
        .push(0) // retOffset
        .push(32) // argSize
        .push(0) // argOffset
        .push(0) // value
        .push(systemContractAddress) // address
        .push(757575) // gas
        .op(OpCode.CALL)
        .op(OpCode.POP) // clean stack
        .compile();
  }

  static Bytes byteCodeCallingBeaconRootSystemAccountFromCallData(
      ChainConfig chainConfig, Address systemContractAddress) {
    return BytecodeCompiler.newProgram(chainConfig)
        // prepare memory with argument
        .op(OpCode.CALLDATASIZE) // size@
        .push(0) // source offset
        .push(0) // destOffset
        .op(OpCode.CALLDATACOPY)

        // call system contract
        .push(0) // retSize
        .push(0) // retOffset
        .push(32) // argSize
        .push(0) // argOffset
        .push(0) // value
        .push(systemContractAddress) // address
        .push(757575) // gas
        .op(OpCode.CALL)
        .compile();
  }
}
