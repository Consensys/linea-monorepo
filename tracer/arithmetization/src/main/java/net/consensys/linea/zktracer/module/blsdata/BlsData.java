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

package net.consensys.linea.zktracer.module.blsdata;

import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.CountingOnlyModule;
import net.consensys.linea.zktracer.container.module.IncrementingModule;
import net.consensys.linea.zktracer.container.module.OperationListModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;

@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class BlsData implements OperationListModule<BlsDataOperation> {
  private final Wcp wcp;
  private final IncrementingModule pointEvaluationEffectiveCall;
  private final IncrementingModule pointEvaluationFailureCall;
  private final IncrementingModule blsG1AddEffectiveCall;
  private final IncrementingModule blsG1MsmEffectiveCall;
  private final IncrementingModule blsG2AddEffectiveCall;
  private final IncrementingModule blsG2MsmEffectiveCall;
  private final CountingOnlyModule blsPairingCheckMillerLoops;
  private final IncrementingModule blsPairingCheckFinalExponentiations;
  private final IncrementingModule blsG1MapFpToG1EffectiveCall;
  private final IncrementingModule blsG1MapFp2ToG2EffectiveCall;
  private final IncrementingModule blsC1MembershipCalls;
  private final IncrementingModule blsC2MembershipCalls;
  private final IncrementingModule blsG1MembershipCalls;
  private final IncrementingModule blsG2MembershipCalls;

  private final ModuleOperationStackedList<BlsDataOperation> operations =
      new ModuleOperationStackedList<>();
  @Getter private BlsDataOperation blsDataOperation;

  @Override
  public String moduleKey() {
    return "BLS_DATA";
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders(Trace trace) {
    return trace.blsdata().headers(this.lineCount());
  }

  @Override
  public int spillage(Trace trace) {
    return trace.blsdata().spillage();
  }

  @Override
  public void commit(Trace trace) {
    int stamp = 0;
    long previousId = 0;
    for (BlsDataOperation op : operations.getAll()) {
      op.trace(trace.blsdata(), ++stamp, previousId);
      previousId = op.id();
    }
  }

  public void callBls(
      final int id,
      final PrecompileScenarioFragment.PrecompileFlag precompileFlag,
      final Bytes callData,
      final Bytes returnData,
      final boolean successBit) {
    blsDataOperation =
        BlsDataOperation.of(wcp, id, precompileFlag, callData, returnData, successBit);
    operations.add(blsDataOperation);

    if (!blsDataOperation.mint()) {
      // If data are detected to be malformed internally, all limits are implicitly 0
      switch (blsDataOperation.precompileFlag()) {
        case PRC_POINT_EVALUATION -> {
          if (blsDataOperation.wnon()) {
            pointEvaluationEffectiveCall.updateTally(1);
          } else if (blsDataOperation.mext()) {
            pointEvaluationFailureCall.updateTally(1);
          }
        }
        case PRC_BLS_G1_ADD -> {
          if (blsDataOperation.wnon()) {
            blsG1AddEffectiveCall.updateTally(1);
          } else if (blsDataOperation.mext()) {
            blsC1MembershipCalls.updateTally(1);
          }
        }
        case PRC_BLS_G1_MSM -> {
          if (blsDataOperation.wnon()) {
            blsG1MsmEffectiveCall.updateTally(1);
          } else if (blsDataOperation.mext()) {
            blsG1MembershipCalls.updateTally(1);
          }
        }
        case PRC_BLS_G2_ADD -> {
          if (blsDataOperation.wnon()) {
            blsG2AddEffectiveCall.updateTally(1);
          } else if (blsDataOperation.mext()) {
            blsC2MembershipCalls.updateTally(1);
          }
        }
        case PRC_BLS_G2_MSM -> {
          if (blsDataOperation.wnon()) {
            blsG2MsmEffectiveCall.updateTally(1);
          } else if (blsDataOperation.mext()) {
            blsG2MembershipCalls.updateTally(1);
          }
        }
        case PRC_BLS_PAIRING_CHECK -> {
          if (blsDataOperation.wtrv() || blsDataOperation.wnon()) {
            /*
            G1  | G2  | Circuit
            P   | inf | G1 membership
            inf | Q   | G2 membership
            inf | inf | none
            */

            blsG1MembershipCalls.updateTally(blsDataOperation.trivialPopDueToG2PointCounter());
            blsG2MembershipCalls.updateTally(blsDataOperation.trivialPopDueToG1PointCounter());
            if (blsDataOperation.wnon()) {
              blsPairingCheckMillerLoops.updateTally(blsDataOperation.nontrivialPopCounter());
              blsPairingCheckFinalExponentiations.updateTally(1);
            }
          } else if (blsDataOperation.mext()) {
            if (blsDataOperation.firstPointNotInSubgroupIsSmall()) {
              blsG1MembershipCalls.updateTally(1);
              blsG2MembershipCalls.updateTally(0);
            } else {
              blsG1MembershipCalls.updateTally(0);
              blsG2MembershipCalls.updateTally(1);
            }
          }
        }
        case PRC_BLS_MAP_FP_TO_G1 -> {
          if (blsDataOperation.wnon()) {
            blsG1MapFpToG1EffectiveCall.updateTally(1);
          }
        }
        case PRC_BLS_MAP_FP2_TO_G2 -> {
          if (blsDataOperation.wnon()) {
            blsG1MapFp2ToG2EffectiveCall.updateTally(1);
          }
        }
      }
    }
  }
}
