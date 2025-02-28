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
package net.consensys.linea.zktracer.module.hub.fragment.imc.oob;

import lombok.AllArgsConstructor;
import lombok.Getter;
import net.consensys.linea.zktracer.Trace;

@Getter
@AllArgsConstructor
public enum OobInstruction {
  OOB_INST_BLAKE_CDS(Trace.OOB_INST_BLAKE_CDS),
  OOB_INST_BLAKE_PARAMS(Trace.OOB_INST_BLAKE_PARAMS),
  OOB_INST_CALL(Trace.OOB_INST_CALL),
  OOB_INST_CDL(Trace.OOB_INST_CDL),
  OOB_INST_CREATE(Trace.OOB_INST_CREATE),
  OOB_INST_DEPLOYMENT(Trace.OOB_INST_DEPLOYMENT),
  OOB_INST_ECADD(Trace.OOB_INST_ECADD),
  OOB_INST_ECMUL(Trace.OOB_INST_ECMUL),
  OOB_INST_ECPAIRING(Trace.OOB_INST_ECPAIRING),
  OOB_INST_ECRECOVER(Trace.OOB_INST_ECRECOVER),
  OOB_INST_IDENTITY(Trace.OOB_INST_IDENTITY),
  OOB_INST_JUMP(Trace.OOB_INST_JUMP),
  OOB_INST_JUMPI(Trace.OOB_INST_JUMPI),
  OOB_INST_MODEXP_CDS(Trace.OOB_INST_MODEXP_CDS),
  OOB_INST_MODEXP_EXTRACT(Trace.OOB_INST_MODEXP_EXTRACT),
  OOB_INST_MODEXP_LEAD(Trace.OOB_INST_MODEXP_LEAD),
  OOB_INST_MODEXP_PRICING(Trace.OOB_INST_MODEXP_PRICING),
  OOB_INST_MODEXP_XBS(Trace.OOB_INST_MODEXP_XBS),
  OOB_INST_RDC(Trace.OOB_INST_RDC),
  OOB_INST_RIPEMD(Trace.OOB_INST_RIPEMD),
  OOB_INST_SHA2(Trace.OOB_INST_SHA2),
  OOB_INST_SSTORE(Trace.OOB_INST_SSTORE),
  OOB_INST_XCALL(Trace.OOB_INST_XCALL);

  public boolean isEvmInstruction() {
    return this.isAnyOf(
        OOB_INST_CALL,
        OOB_INST_CDL,
        OOB_INST_CREATE,
        OOB_INST_DEPLOYMENT,
        OOB_INST_JUMP,
        OOB_INST_JUMPI,
        OOB_INST_RDC,
        OOB_INST_SSTORE,
        OOB_INST_XCALL);
  }

  public boolean isCommonPrecompile() {
    return this.isAnyOf(
        OOB_INST_ECRECOVER,
        OOB_INST_SHA2,
        OOB_INST_RIPEMD,
        OOB_INST_IDENTITY,
        OOB_INST_ECADD,
        OOB_INST_ECMUL,
        OOB_INST_ECPAIRING);
  }

  public boolean isModexp() {
    return this.isAnyOf(
        OOB_INST_MODEXP_CDS,
        OOB_INST_MODEXP_EXTRACT,
        OOB_INST_MODEXP_LEAD,
        OOB_INST_MODEXP_PRICING,
        OOB_INST_MODEXP_XBS);
  }

  public boolean isBlake() {
    return this.isAnyOf(OOB_INST_BLAKE_PARAMS, OOB_INST_BLAKE_CDS);
  }

  public boolean isPrecompile() {
    return this.isCommonPrecompile() || this.isModexp() || this.isBlake();
  }

  public boolean isAnyOf(OobInstruction... oobInstructions) {
    for (OobInstruction oobInstruction : oobInstructions) {
      if (this == oobInstruction) return true;
    }
    return false;
  }

  private final int value;
}
