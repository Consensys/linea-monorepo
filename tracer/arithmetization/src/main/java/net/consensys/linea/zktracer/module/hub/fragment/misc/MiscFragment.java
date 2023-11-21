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

package net.consensys.linea.zktracer.module.hub.fragment.misc;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Signals;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.ExpSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.MmuSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.MxpSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.oob.CalldataloadSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.oob.JumpSubFragment;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

public class MiscFragment implements TraceFragment, PostExecDefer {
  private final Signals signals;
  private final List<TraceSubFragment> subFragments = new ArrayList<>();

  public MiscFragment(Hub hub, MessageFrame frame) {
    this.signals = hub.pch().signals().snapshot();

    // TODO: the rest
    if (this.signals.mmu()) {
      this.subFragments.add(new MmuSubFragment(hub, frame));
    }

    if (this.signals.mxp()) {
      this.subFragments.add(MxpSubFragment.build(hub));
    }

    if (this.signals.oob()) {
      switch (hub.currentFrame().opCode()) {
        case JUMP, JUMPI -> this.subFragments.add(JumpSubFragment.build(hub, frame));
        case CALLDATALOAD -> this.subFragments.add(CalldataloadSubFragment.build(hub, frame));
        case CALL, DELEGATECALL, STATICCALL, CALLCODE -> {}
        case CREATE, CREATE2 -> {}
        case SSTORE -> {}
        case RETURN -> {}
        default -> throw new IllegalArgumentException("unexpected opcode for OoB");
      }
    }

    if (this.signals.exp()) {
      this.subFragments.add(new ExpSubFragment(EWord.of(frame.getStackItem(1))));
    }
  }

  @Override
  public Trace trace(Trace trace) {
    trace
        .peekAtMiscellaneous(true)
        .pMiscellaneousMmuFlag(this.signals.mmu())
        .pMiscellaneousMxpFlag(this.signals.mxp())
        .pMiscellaneousOobFlag(this.signals.oob())
        .pMiscellaneousStpFlag(this.signals.stp())
        .pMiscellaneousExpFlag(this.signals.exp());

    for (TraceSubFragment subFragment : this.subFragments) {
      subFragment.trace(trace);
    }

    return trace;
  }

  @Override
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    for (TraceSubFragment f : this.subFragments) {
      if (f instanceof MmuSubFragment mmuSubFragment) {
        mmuSubFragment.runPostExec(hub, frame, operationResult);
      }
    }
  }
}
