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
package net.consensys.linea.zktracer.module.hub.fragment;

import static net.consensys.linea.zktracer.module.hub.Trace.*;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;

@RequiredArgsConstructor
public class DomSubStampsSubFragment implements TraceSubFragment {

  public enum DomSubType {
    STANDARD,
    REVERTS_WITH_CURRENT,
    REVERTS_WITH_CHILD,
    SELFDESTRUCT
  }

  final DomSubType type;
  final int hubStamp;
  final int domOffset;
  final int subOffset;
  final int revertStamp;
  final int childRevertStamp;
  final int transactionEndStamp;

  public static DomSubStampsSubFragment standardDomSubStamps(final int h, final int domOffset) {
    return new DomSubStampsSubFragment(DomSubType.STANDARD, h, domOffset, 0, 0, 0, 0);
  }

  // TODO: to be use with care, as stamps might have changed
  // TODO: remove altogether
  public static DomSubStampsSubFragment revertWithCurrentDomSubStamps(
      final Hub hub, final int subOffset) {
    return new DomSubStampsSubFragment(
        DomSubType.REVERTS_WITH_CURRENT,
        hub.stamp(),
        0,
        subOffset,
        hub.callStack().currentCallFrame().revertStamp(),
        0,
        0);
  }

  public static DomSubStampsSubFragment revertWithCurrentDomSubStamps(
      final int h, final int revertStamp, final int subOffset) {
    return new DomSubStampsSubFragment(
        DomSubType.REVERTS_WITH_CURRENT, h, 0, subOffset, revertStamp, 0, 0);
  }

  // TODO: to be use with care, as stamps might have changed
  public static DomSubStampsSubFragment revertsWithChildDomSubStamps(
      final Hub hub, final CallFrame child, final int subOffset) {
    return new DomSubStampsSubFragment(
        DomSubType.REVERTS_WITH_CHILD, hub.stamp(), 0, subOffset, 0, child.revertStamp(), 0);
  }

  public static DomSubStampsSubFragment revertsWithChildDomSubStamps(
      final int h, final int childRevertStamp, final int subOffset) {
    return new DomSubStampsSubFragment(
        DomSubType.REVERTS_WITH_CHILD, h, 0, subOffset, 0, childRevertStamp, 0);
  }

  // TODO: to be use with care, as stamps might have changed
  public static DomSubStampsSubFragment selfdestructDomSubStamps(Hub hub, int hubStamp) {
    return new DomSubStampsSubFragment(
        DomSubType.SELFDESTRUCT,
        hubStamp,
        0,
        0,
        0,
        0,
        hub.txStack().current().getHubStampTransactionEnd());
  }

  public int domStamp() {
    switch (type) {
      case STANDARD -> {
        return MULTIPLIER___DOM_SUB_STAMPS * hubStamp + domOffset;
      }
      case REVERTS_WITH_CURRENT -> {
        return MULTIPLIER___DOM_SUB_STAMPS * revertStamp + DOM_SUB_STAMP_OFFSET___REVERT;
      }
      case REVERTS_WITH_CHILD -> {
        return MULTIPLIER___DOM_SUB_STAMPS * childRevertStamp + DOM_SUB_STAMP_OFFSET___REVERT;
      }
      case SELFDESTRUCT -> {
        return MULTIPLIER___DOM_SUB_STAMPS * transactionEndStamp
            + DOM_SUB_STAMP_OFFSET___SELFDESTRUCT;
      }
      default -> throw new RuntimeException("DomSubType not supported");
    }
  }

  public int subStamp() {
    switch (type) {
      case STANDARD -> {
        return subOffset;
      }
      case REVERTS_WITH_CURRENT, REVERTS_WITH_CHILD, SELFDESTRUCT -> {
        return MULTIPLIER___DOM_SUB_STAMPS * hubStamp + subOffset;
      }
      default -> throw new RuntimeException("DomSubType not supported");
    }
  }

  @Override
  public Trace trace(Trace trace) {
    return trace.domStamp(this.domStamp()).subStamp(this.subStamp());
  }
}
