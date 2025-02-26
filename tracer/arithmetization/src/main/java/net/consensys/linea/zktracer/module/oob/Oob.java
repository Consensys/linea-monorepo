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

package net.consensys.linea.zktracer.module.oob;

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.*;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.wcp.Wcp;

/** Implementation of a {@link Module} for out of bounds. */
@RequiredArgsConstructor
@Accessors(fluent = true)
public class Oob implements OperationSetModule<OobOperation> {

  private final Hub hub;
  private final Add add;
  private final Mod mod;
  private final Wcp wcp;

  @Getter
  private final ModuleOperationStackedSet<OobOperation> operations =
      new ModuleOperationStackedSet<>();

  @Override
  public String moduleKey() {
    return "OOB";
  }

  public void call(OobCall oobCall) {
    final OobOperation oobOperation =
        new OobOperation(oobCall, hub.messageFrame(), add, mod, wcp, hub);
    operations.add(oobOperation);
  }

  final void traceOperation(final OobOperation oobOperation, int stamp, Trace trace) {
    final int nRows = oobOperation.nRows();
    final OobInstruction oobInstruction = oobOperation.oobCall.oobInstruction;

    for (int ct = 0; ct < nRows; ct++) {
      trace = oobOperation.getOobCall().trace(trace);

      // Note: if a value is bigger than 128, do not use Bytes.of and use Bytes.ofUnsignedType
      // instead (according to size)
      trace
          .stamp(stamp)
          .ct((short) ct)
          .ctMax((short) oobOperation.ctMax())
          .oobInst(oobInstruction.getValue())
          .isJump(oobInstruction == OOB_INST_JUMP)
          .isJumpi(oobInstruction == OOB_INST_JUMPI)
          .isRdc(oobInstruction == OOB_INST_RDC)
          .isCdl(oobInstruction == OOB_INST_CDL)
          .isXcall(oobInstruction == OOB_INST_XCALL)
          .isCall(oobInstruction == OOB_INST_CALL)
          .isCreate(oobInstruction == OOB_INST_CREATE)
          .isSstore(oobInstruction == OOB_INST_SSTORE)
          .isDeployment(oobInstruction == OOB_INST_DEPLOYMENT)
          .isEcrecover(oobInstruction == OOB_INST_ECRECOVER)
          .isSha2(oobInstruction == OOB_INST_SHA2)
          .isRipemd(oobInstruction == OOB_INST_RIPEMD)
          .isIdentity(oobInstruction == OOB_INST_IDENTITY)
          .isEcadd(oobInstruction == OOB_INST_ECADD)
          .isEcmul(oobInstruction == OOB_INST_ECMUL)
          .isEcpairing(oobInstruction == OOB_INST_ECPAIRING)
          .isBlake2FCds(oobInstruction == OOB_INST_BLAKE_CDS)
          .isBlake2FParams(oobInstruction == OOB_INST_BLAKE_PARAMS)
          .isModexpCds(oobInstruction == OOB_INST_MODEXP_CDS)
          .isModexpXbs(oobInstruction == OOB_INST_MODEXP_XBS)
          .isModexpLead(oobInstruction == OOB_INST_MODEXP_LEAD)
          .isModexpPricing(oobInstruction == OOB_INST_MODEXP_PRICING)
          .isModexpExtract(oobInstruction == OOB_INST_MODEXP_EXTRACT)
          .addFlag(oobOperation.getAddFlag()[ct])
          .modFlag(oobOperation.getModFlag()[ct])
          .wcpFlag(oobOperation.getWcpFlag()[ct])
          .outgoingInst(oobOperation.getOutgoingInst()[ct])
          .outgoingData1(bigIntegerToBytes(oobOperation.getOutgoingData1()[ct]))
          .outgoingData2(bigIntegerToBytes(oobOperation.getOutgoingData2()[ct]))
          .outgoingData3(bigIntegerToBytes(oobOperation.getOutgoingData3()[ct]))
          .outgoingData4(bigIntegerToBytes(oobOperation.getOutgoingData4()[ct]))
          .outgoingResLo(bigIntegerToBytes(oobOperation.getOutgoingResLo()[ct]))
          .validateRow();
    }
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    Trace trace = new Trace(buffers);
    int stamp = 0;
    for (OobOperation op : operations.getAll()) {
      traceOperation(op, ++stamp, trace);
    }
  }

  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }
}
