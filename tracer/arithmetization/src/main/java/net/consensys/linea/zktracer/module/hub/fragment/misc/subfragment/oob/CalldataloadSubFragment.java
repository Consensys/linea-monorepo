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

package net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.oob;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

public record CalldataloadSubFragment(EWord readOffset, EWord calldataSize)
    implements TraceSubFragment {
  private static final Bytes CALLDATALOAD = Bytes.of(OpCode.CALLDATALOAD.byteValue());

  public static CalldataloadSubFragment build(Hub hub, MessageFrame frame) {
    return new CalldataloadSubFragment(
        EWord.of(frame.getStackItem(0)), EWord.of(hub.currentFrame().callData().size()));
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscellaneousOobOutgoingData1(this.readOffset.hi())
        .pMiscellaneousOobOutgoingData2(this.readOffset().lo())
        .pMiscellaneousOobOutgoingData5(this.calldataSize)
        .pMiscellaneousOobInst(CALLDATALOAD)
        .pMiscellaneousOobEvent1(this.readOffset.greaterOrEqualThan(this.calldataSize));
  }
}
