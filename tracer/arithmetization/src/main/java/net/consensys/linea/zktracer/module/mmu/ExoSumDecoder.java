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

package net.consensys.linea.zktracer.module.mmu;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpData;
import net.consensys.linea.zktracer.module.ecdata.EcData;
import net.consensys.linea.zktracer.module.mmu.values.HubToMmuValues;
import net.consensys.linea.zktracer.module.rlptxn.RlpTxn;
import net.consensys.linea.zktracer.module.rlptxrcpt.RlpTxnRcpt;
import net.consensys.linea.zktracer.module.romlex.RomLex;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import org.apache.tuweni.bytes.Bytes;

@Getter
@Accessors(fluent = true)
@RequiredArgsConstructor
public class ExoSumDecoder {
  private final CallStack callStack;
  private final RomLex romLex;
  private final RlpTxn rlpTxn;
  private final RlpTxnRcpt rlpTxnRcpt;
  private final EcData ecData;
  private final BlakeModexpData blakeModexpData;

  private boolean exoIsRom;
  private boolean exoIsBlake2fModexp;
  private boolean exoIsEcData;
  private boolean exoIsRipSha;
  private boolean exoIsKeccak;
  private boolean exoIsLog;
  private boolean exoIsTxcd;

  public Bytes getExoBytes(final HubToMmuValues hubToMmuValues, final int exoId) {
    exoIsRom = hubToMmuValues.exoIsRom();
    exoIsBlake2fModexp = hubToMmuValues.exoIsBlake2fModexp();
    exoIsEcData = hubToMmuValues.exoIsEcData();
    exoIsRipSha = hubToMmuValues.exoIsRipSha();
    exoIsKeccak = hubToMmuValues.exoIsKeccak();
    exoIsLog = hubToMmuValues.exoIsLog();
    exoIsTxcd = hubToMmuValues.exoIsTxcd();

    if (exoIsRom) {
      return this.romLex.sortedChunks().get(exoId - 1).byteCode().copy();
    }

    if (exoIsTxcd) {
      return this.rlpTxn.chunkList.get(exoId - 1).tx().getPayload();
    }

    if (exoIsLog) {
      return this.rlpTxnRcpt.getLogDataByAbsLogNumber(exoId);
    }

    if (exoIsEcData) {
      return Bytes.EMPTY;
      // TODO
    }

    if (exoIsRipSha) {
      return Bytes.EMPTY;
      // TODO
    }

    if (exoIsBlake2fModexp) {
      return this.blakeModexpData.getInputDataByIdAndPhase(exoId, hubToMmuValues.phase());
    }

    if (exoIsKeccak) {
      return Bytes.EMPTY;
      // TODO use hubToMmuValues.auxId()
    }

    throw new IllegalArgumentException("No exo flag set, can't retrieve exo bytes");
  }
}
