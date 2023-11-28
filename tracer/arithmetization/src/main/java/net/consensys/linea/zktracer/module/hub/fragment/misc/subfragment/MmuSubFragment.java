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

package net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment;

import java.math.BigInteger;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.operation.Operation;

@Slf4j
public class MmuSubFragment implements TraceSubFragment, PostExecDefer {
  Bytes stackValue = Bytes.EMPTY;
  Bytes offset1 = Bytes.EMPTY;
  Bytes offset2 = Bytes.EMPTY;
  byte opCode;
  int param1 = 0;
  int param2 = 0;
  int returner = 0;
  boolean info = false;
  long referenceOffset = 0;
  long referenceSize = 0;
  int exoSum = 0;
  int size = 0;

  public MmuSubFragment(Hub hub, MessageFrame frame) {
    final OpCode opCode = hub.currentFrame().opCode();

    switch (opCode) {
      case SHA3 -> {
        offset1 = frame.getStackItem(0).copy();
        param1 = 0; // TODO: hash info stamp
        size = Words.clampedToInt(Words.clampedToLong(frame.getStackItem(1)));
        exoSum = 0; // TODO:
      }
      case CALLDATALOAD -> {
        this.param1 = hub.tx().number();
        this.info = hub.callStack().depth() == 1;
        this.referenceOffset = hub.currentFrame().callDataPointer().offset();
        this.referenceSize = hub.currentFrame().callDataPointer().length();
        this.offset1 = frame.getStackItem(0).copy();
      }
      case MSTORE, MSTORE8 -> {
        this.offset1 = frame.getStackItem(0).copy();
        this.stackValue = frame.getStackItem(1).copy();
      }
      case MLOAD -> this.offset1 = frame.getStackItem(0).copy();
      default -> log.info("MMU not yet implemented for this opcode");
    }

    this.opCode = opCode.byteValue();
  }

  @Override
  public void runPostExec(
      final Hub hub, final MessageFrame frame, final Operation.OperationResult operationResult) {
    switch (hub.opCode()) {
      case MLOAD, CALLDATALOAD -> {
        this.stackValue = frame.getStackItem(0).copy();
      }
      default -> {}
    }
  }

  @Override
  public Trace trace(Trace trace) {
    final EWord eOffset1 = EWord.of(this.offset1);
    final EWord eOffset2 = EWord.of(this.offset2);
    final EWord eStackValue = EWord.of(this.stackValue);

    return trace
        .pMiscellaneousMmuInst(BigInteger.valueOf(this.opCode))
        .pMiscellaneousMmuParam1(BigInteger.valueOf(this.param1))
        .pMiscellaneousMmuParam2(BigInteger.valueOf(this.param2))
        .pMiscellaneousMmuReturner(BigInteger.valueOf(this.returner))
        .pMiscellaneousMmuInfo(this.info)
        .pMiscellaneousMmuRefOffset(BigInteger.valueOf(this.referenceOffset))
        .pMiscellaneousMmuRefSize(BigInteger.valueOf(this.referenceSize))
        .pMiscellaneousMmuOffset1Lo(eOffset1.loBigInt())
        .pMiscellaneousMmuOffset2Hi(eOffset2.hiBigInt())
        .pMiscellaneousMmuOffset2Lo(eOffset2.loBigInt())
        .pMiscellaneousMmuSize(BigInteger.valueOf(this.size))
        .pMiscellaneousMmuStackValHi(eStackValue.hiBigInt())
        .pMiscellaneousMmuStackValLo(eStackValue.loBigInt())
        .pMiscellaneousMmuExoSum(BigInteger.valueOf(this.exoSum));
  }
}
