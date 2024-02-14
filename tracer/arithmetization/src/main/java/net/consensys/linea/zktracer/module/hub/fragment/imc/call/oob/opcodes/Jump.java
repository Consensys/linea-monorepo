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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.opcodes;

import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_jump;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_jumpi;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public final class Jump implements OobCall {
  private final EWord targetPc;
  private final EWord jumpCondition;
  private final int codeSize;
  private final int oobInst;
  private final boolean event1;
  private final boolean event2;

  public Jump(Hub hub, MessageFrame frame) {
    long targetPc = Words.clampedToLong(frame.getStackItem(0));
    final boolean invalidDestination = frame.getCode().isJumpDestInvalid((int) targetPc);

    long jumpCondition = 0;
    switch (hub.currentFrame().opCode()) {
      case JUMP -> {
        this.oobInst = OOB_INST_jump;
        this.event1 = invalidDestination;
      }
      case JUMPI -> {
        this.oobInst = OOB_INST_jumpi;
        jumpCondition = Words.clampedToLong(frame.getStackItem(1));
        this.event1 = (jumpCondition != 0) && invalidDestination;
      }
      default -> throw new IllegalArgumentException("Unexpected opcode");
    }

    this.targetPc = EWord.of(targetPc);
    this.jumpCondition = EWord.of(jumpCondition);
    this.codeSize = frame.getWorldUpdater().get(hub.currentFrame().codeAddress()).getCode().size();
    this.event2 = jumpCondition > 0;
  }

  @Override
  public int oobInstruction() {
    return this.oobInst;
  }

  @Override
  public Bytes data(OobDataChannel i) {
    return switch (i) {
      case DATA_1 -> this.targetPc.hi();
      case DATA_2 -> this.targetPc.lo();
      case DATA_3 -> this.jumpCondition.hi();
      case DATA_4 -> this.jumpCondition.lo();
      case DATA_5 -> Bytes.ofUnsignedInt(this.codeSize);
      case DATA_7 -> this.event1 ? Bytes.of(1) : Bytes.EMPTY;
      case DATA_8 -> this.event2 ? Bytes.of(1) : Bytes.EMPTY;
      default -> Bytes.EMPTY;
    };
  }
}
