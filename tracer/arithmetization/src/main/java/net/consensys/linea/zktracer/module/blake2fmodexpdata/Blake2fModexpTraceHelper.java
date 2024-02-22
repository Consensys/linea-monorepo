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

package net.consensys.linea.zktracer.module.blake2fmodexpdata;

import java.util.Map;
import java.util.function.BiConsumer;

import lombok.Builder;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.commons.lang3.function.TriFunction;
import org.apache.tuweni.bytes.Bytes;

@Builder
@Accessors(fluent = true)
public class Blake2fModexpTraceHelper {
  private final Bytes currentHubStamp;
  @Getter private int prevHubStamp;
  private final int startPhaseIndex;
  private final int endPhaseIndex;

  private final Map<Integer, PhaseInfo> phaseInfoMap;
  private final Trace trace;
  private final UnsignedByte stampByte;
  private final BiConsumer<Integer, Integer> traceLimbConsumer;
  private final TriFunction<PhaseInfo, Integer, Integer, Integer> currentRowIndexFunction;
  private UnsignedByte[] hubStampDiffBytes;

  void trace() {
    boolean[] phaseFlags = new boolean[7];

    for (int phaseIndex = startPhaseIndex; phaseIndex <= endPhaseIndex; phaseIndex++) {
      phaseFlags[phaseIndex - 1] = true;

      final PhaseInfo phaseInfo = phaseInfoMap.get(phaseIndex);

      final int indexMax = phaseInfo.indexMax();
      for (int index = 0; index <= indexMax; index++) {
        int rowIndex = currentRowIndexFunction.apply(phaseInfo, phaseIndex, index);

        trace
            .phase(UnsignedByte.of(phaseInfo.id()))
            .deltaByte(hubStampDiffBytes[rowIndex])
            .id(currentHubStamp)
            .index(UnsignedByte.of(index))
            .indexMax(UnsignedByte.of(indexMax))
            .stamp(stampByte);

        traceLimbConsumer.accept(rowIndex, phaseIndex);

        trace
            .isModexpBase(phaseFlags[0])
            .isModexpExponent(phaseFlags[1])
            .isModexpModulus(phaseFlags[2])
            .isModexpResult(phaseFlags[3])
            .isBlakeData(phaseFlags[4])
            .isBlakeParams(phaseFlags[5])
            .isBlakeResult(phaseFlags[6])
            .validateRow();
      }

      phaseFlags[phaseIndex - 1] = false;
    }

    prevHubStamp = currentHubStamp.toInt();
  }
}
