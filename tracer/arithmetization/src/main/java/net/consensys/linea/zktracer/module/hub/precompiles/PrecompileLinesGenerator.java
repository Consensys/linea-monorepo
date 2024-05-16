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

package net.consensys.linea.zktracer.module.hub.precompiles;

import java.util.ArrayList;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.exp.ModexpLogOperation;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.exp.ExpCallForModexpLogComputation;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.Blake2FPrecompile1;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.Blake2FPrecompile2;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.EcAdd;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.EcMul;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.EcPairing;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.EcRecover;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.Identity;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.ModexpCds;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.ModexpExtract;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.ModexpLead;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.ModexpPricing;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.ModexpXbs;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.RipeMd160;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles.Sha2;
import net.consensys.linea.zktracer.types.EWord;

@RequiredArgsConstructor
public class PrecompileLinesGenerator {
  public static List<TraceFragment> generateFor(final Hub hub, final PrecompileInvocation p) {
    final List<TraceFragment> r = new ArrayList<>(10);
    switch (p.precompile()) {
      case EC_RECOVER -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcRecover(p)));
        } else {
          final boolean recoverySuccessful =
              ((EcRecoverMetadata) p.metadata()).recoverySuccessful();

          r.add(
              ImcFragment.empty(hub)
                  .callOob(new EcRecover(p))
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcRecover(hub, p, recoverySuccessful, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcRecover(hub, p, recoverySuccessful, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcRecover(hub, p, recoverySuccessful, 2)));
        }
      }
      case SHA2_256 -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new Sha2(p)));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Sha2(p))
                  .callMmu(
                      p.callDataSource().isEmpty() ? MmuCall.nop() : MmuCall.forSha2(hub, p, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty() ? MmuCall.nop() : MmuCall.forSha2(hub, p, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty() ? MmuCall.nop() : MmuCall.forSha2(hub, p, 2)));
        }
      }
      case RIPEMD_160 -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new RipeMd160(p)));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Sha2(p))
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forRipeMd160(hub, p, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forRipeMd160(hub, p, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forRipeMd160(hub, p, 2)));
        }
      }
      case IDENTITY -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new Identity(p)));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Identity(p))
                  .callMmu(MmuCall.forIdentity(hub, p, 0)));
          r.add(ImcFragment.empty(hub).callMmu(MmuCall.forIdentity(hub, p, 1)));
        }
      }
      case MODEXP -> {
        final ModExpMetadata m = (ModExpMetadata) p.metadata();
        final int bbsInt = m.bbs().toUnsignedBigInteger().intValueExact();
        final int ebsInt = m.ebs().toUnsignedBigInteger().intValueExact();
        final int mbsInt = m.mbs().toUnsignedBigInteger().intValueExact();

        r.add(
            ImcFragment.empty(hub).callOob(new ModexpCds(p.requestedReturnDataTarget().length())));
        r.add(
            ImcFragment.empty(hub)
                .callOob(new ModexpXbs(m.bbs(), EWord.ZERO, false))
                .callMmu(m.extractBbs() ? MmuCall.forModExp(hub, p, 2) : MmuCall.nop()));
        r.add(
            ImcFragment.empty(hub)
                .callOob(new ModexpXbs(m.ebs(), EWord.ZERO, false))
                .callMmu(m.extractEbs() ? MmuCall.forModExp(hub, p, 3) : MmuCall.nop()));
        r.add(
            ImcFragment.empty(hub)
                .callOob(new ModexpXbs(m.mbs(), m.bbs(), true))
                .callMmu(m.extractEbs() ? MmuCall.forModExp(hub, p, 4) : MmuCall.nop()));
        final ImcFragment line5 =
            ImcFragment.empty(hub)
                .callOob(new ModexpLead(bbsInt, p.callDataSource().length(), ebsInt))
                .callMmu(m.loadRawLeadingWord() ? MmuCall.forModExp(hub, p, 5) : MmuCall.nop());
        if (m.loadRawLeadingWord()) {
          line5.callExp(
              new ExpCallForModexpLogComputation(
                  m.rawLeadingWord(),
                  Math.min((int) (p.callDataSource().length() - 96 - bbsInt), 32),
                  Math.min(ebsInt, 32)));
        }
        r.add(line5);
        r.add(
            ImcFragment.empty(hub)
                .callOob(
                    new ModexpPricing(
                        p,
                        m.loadRawLeadingWord()
                            ? ModexpLogOperation.LeadLogTrimLead.fromArgs(
                                    m.rawLeadingWord(),
                                    Math.min((int) (p.callDataSource().length() - 96 - bbsInt), 32),
                                    Math.min(ebsInt, 32))
                                .leadLog()
                            : 0,
                        Math.max(mbsInt, bbsInt))));

        if (p.ramFailure()) {
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new ModexpExtract(p.callDataSource().length(), bbsInt, ebsInt, mbsInt))
                  .callMmu(m.extractModulus() ? MmuCall.forModExp(hub, p, 7) : MmuCall.nop()));

          if (m.extractModulus()) {
            r.add(ImcFragment.empty(hub).callMmu(MmuCall.forModExp(hub, p, 8)));
            r.add(ImcFragment.empty(hub).callMmu(MmuCall.forModExp(hub, p, 9)));
            r.add(ImcFragment.empty(hub).callMmu(MmuCall.forModExp(hub, p, 10)));
          } else {
            for (int i = 0; i < 4; i++) {
              r.add(ImcFragment.empty(hub));
            }
          }

          if (!m.mbs().isZero() && !p.requestedReturnDataTarget().isEmpty()) {
            r.add(ImcFragment.empty(hub).callMmu(MmuCall.forModExp(hub, p, 11)));
          }

          r.add(ContextFragment.providesReturnData(hub.callStack()));
        }
      }
      case EC_ADD -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcAdd(p)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else if (p.ramFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcAdd(p)).callMmu(MmuCall.forEcAdd(hub, p, 0)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else {
          r.add(ImcFragment.empty(hub).callOob(new EcAdd(p)).callMmu(MmuCall.forEcAdd(hub, p, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty() ? MmuCall.nop() : MmuCall.forEcAdd(hub, p, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.requestedReturnDataTarget().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcAdd(hub, p, 2)));
          r.add(ContextFragment.providesReturnData(hub.callStack()));
        }
      }
      case EC_MUL -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcMul(p)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else if (p.ramFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcMul(p)).callMmu(MmuCall.forEcMul(hub, p, 0)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else {
          r.add(ImcFragment.empty(hub).callOob(new EcMul(p)).callMmu(MmuCall.forEcMul(hub, p, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.callDataSource().isEmpty() ? MmuCall.nop() : MmuCall.forEcMul(hub, p, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.requestedReturnDataTarget().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcMul(hub, p, 2)));
          r.add(ContextFragment.providesReturnData(hub.callStack()));
        }
      }
      case EC_PAIRING -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new EcPairing(p)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else if (p.ramFailure()) {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new EcPairing(p))
                  .callMmu(MmuCall.forEcPairing(hub, p, 0)));
          r.add(ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new EcPairing(p))
                  .callMmu(MmuCall.forEcPairing(hub, p, 0)));
          r.add(ImcFragment.empty(hub).callMmu(MmuCall.forEcPairing(hub, p, 1)));
          r.add(
              ImcFragment.empty(hub)
                  .callMmu(
                      p.requestedReturnDataTarget().isEmpty()
                          ? MmuCall.nop()
                          : MmuCall.forEcPairing(hub, p, 2)));
          r.add(ContextFragment.providesReturnData(hub.callStack()));
        }
      }
      case BLAKE2F -> {
        if (p.hubFailure()) {
          r.add(ImcFragment.empty(hub).callOob(new Blake2FPrecompile1(p)));
        } else if (p.ramFailure()) {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Blake2FPrecompile1(p))
                  .callMmu(MmuCall.forBlake2f(hub, p, 0)));
          r.add(ImcFragment.empty(hub).callOob(new Blake2FPrecompile2(p)));
        } else {
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Blake2FPrecompile1(p))
                  .callMmu(MmuCall.forBlake2f(hub, p, 0)));
          r.add(
              ImcFragment.empty(hub)
                  .callOob(new Blake2FPrecompile2(p))
                  .callMmu(MmuCall.forBlake2f(hub, p, 1)));
          r.add(ImcFragment.empty(hub).callMmu(MmuCall.forBlake2f(hub, p, 2)));
          r.add(ImcFragment.empty(hub).callMmu(MmuCall.forBlake2f(hub, p, 3)));
        }
      }
    }

    r.add(
        p.success()
            ? ContextFragment.providesReturnData(hub.callStack())
            : ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
    return r;
  }
}
