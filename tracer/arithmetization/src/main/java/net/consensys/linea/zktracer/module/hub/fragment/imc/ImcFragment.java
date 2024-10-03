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

package net.consensys.linea.zktracer.module.hub.fragment.imc;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;

/**
 * IMCFragments embed data required for Inter-Module Communication, i.e. data that are required to
 * correctly trigger other modules from the Hub.
 */
public class ImcFragment implements TraceFragment {
  /** the list of modules to trigger withing this fragment. */
  private final List<TraceSubFragment> moduleCalls = new ArrayList<>(5);

  private final Hub hub;

  private boolean expIsSet = false;
  private boolean oobIsSet = false;
  private boolean mxpIsSet = false;
  private boolean mmuIsSet = false;
  private boolean stpIsSet = false;

  private ImcFragment(final Hub hub) {
    this.hub = hub;
  }

  /**
   * Create an empty ImcFragment to be filled with specialized methods.
   *
   * @return an empty ImcFragment
   */
  public static ImcFragment empty(final Hub hub) {
    return new ImcFragment(hub);
  }

  /**
   * Create an ImcFragment to be used in the transaction initialization phase.
   *
   * @param hub the execution context
   * @return the ImcFragment for the TxInit phase
   */
  public static ImcFragment forTxInit(final Hub hub) {
    // isdeployment == false
    // non empty calldata
    final TransactionProcessingMetadata currentTx = hub.txStack().current();
    final boolean shouldCopyTxCallData = currentTx.copyTransactionCallData();

    final ImcFragment miscFragment = ImcFragment.empty(hub);

    return shouldCopyTxCallData ? miscFragment.callMmu(MmuCall.txInit(hub)) : miscFragment;
  }

  public ImcFragment callOob(OobCall f) {
    if (oobIsSet) {
      throw new IllegalStateException("OOB already called");
    } else {
      oobIsSet = true;
    }
    hub.oob().call(f);
    moduleCalls.add(f);
    return this;
  }

  public ImcFragment callMmu(MmuCall f) {
    if (mmuIsSet) {
      throw new IllegalStateException("MMU already called");
    } else {
      mmuIsSet = true;
    }
    // Note: the triggering of the MMU is made by the creation of the MmuCall
    moduleCalls.add(f);
    return this;
  }

  public ImcFragment callExp(ExpCall f) {
    if (expIsSet) {
      throw new IllegalStateException("EXP already called");
    } else {
      expIsSet = true;
    }
    hub.exp().call(f);
    moduleCalls.add(f);
    return this;
  }

  public ImcFragment callMxp(MxpCall f) {
    if (mxpIsSet) {
      throw new IllegalStateException("MXP already called");
    } else {
      mxpIsSet = true;
    }
    hub.mxp().call(f);
    moduleCalls.add(f);
    return this;
  }

  public ImcFragment callStp(StpCall f) {
    if (stpIsSet) {
      throw new IllegalStateException("STP already called");
    } else {
      stpIsSet = true;
    }
    hub.stp().call(f);
    moduleCalls.add(f);
    return this;
  }

  @Override
  public Trace trace(Trace trace) {
    trace.peekAtMiscellaneous(true);

    for (TraceSubFragment subFragment : moduleCalls) {
      subFragment.trace(trace, hub.state.stamps());
    }

    return trace;
  }
}
