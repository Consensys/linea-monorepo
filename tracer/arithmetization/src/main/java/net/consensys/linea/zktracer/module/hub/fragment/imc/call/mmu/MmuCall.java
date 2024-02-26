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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu;

import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_BLAKE_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_BLAKE_PARAMETERS;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_BLAKE_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECADD_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECADD_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECMUL_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECMUL_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECRECOVER_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_ECRECOVER_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_MODEXP_BASE;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_MODEXP_EXPONENT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_MODEXP_MODULUS;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_MODEXP_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_PAIRING_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_PAIRING_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_SHA2_256_DATA;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_SHA2_256_RESULT;
import static net.consensys.linea.zktracer.module.hub.Trace.PHASE_TRANSACTION_CALL_DATA;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_ANY_TO_RAM_WITH_PADDING;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_BLAKE;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_EXO_TO_RAM_TRANSPLANTS;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_MLOAD;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_MODEXP_DATA;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_MODEXP_ZERO;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_MSTORE;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_MSTORE8;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_RAM_TO_EXO_WITH_PADDING;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_RAM_TO_RAM_SANS_PADDING;
import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_RIGHT_PADDED_WORD_EXTRACTION;

import java.util.Arrays;

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.CodeCopy;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.Create;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.Create2;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.ExtCodeCopy;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.LogX;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode.ReturnFromDeployment;
import net.consensys.linea.zktracer.module.hub.precompiles.Blake2fMetadata;
import net.consensys.linea.zktracer.module.hub.precompiles.ModExpMetadata;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.internal.Words;

/**
 * This class represents a call to the MMU. However, some MMU calls may have their actual content
 * only definitely defined at tracing time, post conflation. In these cases, subclasses of this
 * class implement this defer mechanism.
 */
@RequiredArgsConstructor
@Setter
@Accessors(fluent = true)
public class MmuCall implements TraceSubFragment {
  protected boolean enabled = true;
  protected int instruction = 0;
  protected int sourceId = 0;
  protected int targetId = 0;
  protected int auxId = 0;
  protected EWord sourceOffset = EWord.ZERO;
  protected EWord targetOffset = EWord.ZERO;
  protected long size = 0;
  protected long referenceOffset = 0;
  protected long referenceSize = 0;
  protected boolean successBit = false;
  protected Bytes limb1 = Bytes.EMPTY;
  protected Bytes limb2 = Bytes.EMPTY;
  protected long phase = 0;

  protected boolean enabled() {
    return this.enabled;
  }

  protected int instruction() {
    return this.instruction;
  }

  protected int sourceId() {
    return this.sourceId;
  }

  protected int targetId() {
    return this.targetId;
  }

  protected int auxId() {
    return this.auxId;
  }

  protected EWord sourceOffset() {
    return this.sourceOffset;
  }

  protected EWord targetOffset() {
    return this.targetOffset;
  }

  protected long size() {
    return this.size;
  }

  protected long referenceOffset() {
    return this.referenceOffset;
  }

  protected long referenceSize() {
    return this.referenceSize;
  }

  protected boolean successBit() {
    return this.successBit;
  }

  protected Bytes limb1() {
    return this.limb1;
  }

  protected Bytes limb2() {
    return this.limb2;
  }

  protected long phase() {
    return this.phase;
  }

  private int exoSum = 0;

  private MmuCall setFlag(int pos) {
    this.exoSum |= 1 >> pos;
    return this;
  }

  final MmuCall setRlpTxn() {
    return this.setFlag(0);
  }

  public final MmuCall setLog() {
    return this.setFlag(1);
  }

  public final MmuCall setRom() {
    return this.setFlag(2);
  }

  public final MmuCall setHash() {
    return this.setFlag(3);
  }

  final MmuCall setRipSha() {
    return this.setFlag(4);
  }

  final MmuCall setBlakeModexp() {
    return this.setFlag(5);
  }

  final MmuCall setEcData() {
    return this.setFlag(6);
  }

  public MmuCall(int instruction) {
    this.instruction = instruction;
  }

  public static MmuCall nop() {
    return new MmuCall();
  }

  public static MmuCall sha3(final Hub hub) {
    return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .auxId(hub.state().stamps().hashInfo())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceSize(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .setHash();
  }

  public static MmuCall callDataLoad(final Hub hub) {
    final long offset = hub.currentFrame().callDataSource().offset();
    final long size = hub.currentFrame().callDataSource().length();

    final long sourceOffset = Words.clampedToLong(hub.messageFrame().getStackItem(0));

    if (sourceOffset >= size) {
      return nop();
    }

    final EWord read =
        EWord.of(
            Bytes.wrap(
                Arrays.copyOfRange(
                    hub.currentFrame().callData().toArray(), (int) offset, (int) (offset + 32))));

    return new MmuCall(MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
        .sourceId(hub.callStack().getById(hub.currentFrame().parentFrame()).contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .referenceOffset(offset)
        .referenceSize(size)
        .limb1(read.hi())
        .limb2(read.lo());
  }

  public static MmuCall callDataCopy(final Hub hub) {
    final MemorySpan callDataSegment = hub.currentFrame().callDataSource();
    return new MmuCall(MMU_INST_ANY_TO_RAM_WITH_PADDING)
        .sourceId(hub.transients().tx().absNumber())
        .targetId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(1)))
        .targetOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(2)))
        .referenceOffset(callDataSegment.offset())
        .referenceSize(callDataSegment.length());
  }

  public static MmuCall codeCopy(final Hub hub) {
    return new CodeCopy(hub);
  }

  public static MmuCall extCodeCopy(final Hub hub) {
    return new ExtCodeCopy(hub);
  }

  public static MmuCall returnDataCopy(final Hub hub) {
    final MemorySpan returnDataSegment = hub.currentFrame().latestReturnDataSource();
    return new MmuCall(MMU_INST_ANY_TO_RAM_WITH_PADDING)
        .sourceId(hub.callStack().getById(hub.currentFrame().currentReturner()).contextNumber())
        .targetId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(1)))
        .targetOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(2)))
        .referenceOffset(returnDataSegment.offset())
        .referenceSize(returnDataSegment.length());
  }

  public static MmuCall mload(final Hub hub) {
    final long offset = Words.clampedToLong(hub.messageFrame().getStackItem(0));
    final EWord loadedValue = EWord.of(hub.messageFrame().shadowReadMemory(offset, 32));
    return new MmuCall(MMU_INST_MLOAD)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .limb1(loadedValue.hi())
        .limb2(loadedValue.lo());
  }

  public static MmuCall mstore(final Hub hub) {
    final EWord storedValue = EWord.of(hub.messageFrame().getStackItem(1));
    return new MmuCall(MMU_INST_MSTORE)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .limb1(storedValue.hi())
        .limb2(storedValue.lo());
  }

  public static MmuCall mstore8(final Hub hub) {
    final EWord storedValue = EWord.of(hub.messageFrame().getStackItem(1));
    return new MmuCall(MMU_INST_MSTORE8)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .limb1(storedValue.hi())
        .limb2(storedValue.lo());
  }

  public static MmuCall log(final Hub hub) {
    return new LogX(hub);
  }

  public static MmuCall create(final Hub hub) {
    return new Create(hub);
  }

  public static MmuCall returnFromDeployment(final Hub hub) {
    return new ReturnFromDeployment(hub);
  }

  public static MmuCall returnFromCall(final Hub hub) {
    return MmuCall.revert(hub);
  }

  public static MmuCall create2(final Hub hub) {
    return new Create2(hub);
  }

  public static MmuCall revert(final Hub hub) {
    return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .targetId(hub.callStack().getById(hub.currentFrame().parentFrame()).contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceOffset(hub.currentFrame().requestedReturnDataTarget().offset())
        .referenceSize(hub.currentFrame().requestedReturnDataTarget().length());
  }

  public static MmuCall txInit(final Hub hub) {
    return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(hub.transients().tx().absNumber())
        .targetId(hub.stamp())
        .size(hub.transients().tx().besuTx().getData().map(Bytes::size).orElse(0))
        .phase(PHASE_TRANSACTION_CALL_DATA)
        .setRlpTxn();
  }

  public static MmuCall forEcRecover(
      final Hub hub, PrecompileInvocation p, boolean recoverySuccessful, int i) {
    Preconditions.checkArgument(i >= 0 && i < 3);

    if (i == 0) {
      return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .size(p.callDataSource().length())
          .referenceSize(128)
          .successBit(recoverySuccessful)
          .phase(PHASE_ECRECOVER_DATA)
          .setEcData();
    } else if (i == 1) {
      if (recoverySuccessful) {
        return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.stamp() + 1)
            .size(32)
            .phase(PHASE_ECRECOVER_RESULT)
            .setEcData();
      } else {
        return nop();
      }
    } else {
      if (recoverySuccessful && !p.requestedReturnDataTarget().isEmpty()) {
        return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.currentFrame().contextNumber())
            .sourceOffset(EWord.ZERO)
            .size(32)
            .referenceOffset(p.requestedReturnDataTarget().offset())
            .referenceSize(p.requestedReturnDataTarget().length());

      } else {
        return nop();
      }
    }
  }

  private static MmuCall forRipeMd160Sha(
      final Hub hub, PrecompileInvocation p, int i, Bytes emptyHi, Bytes emptyLo) {
    Preconditions.checkArgument(i >= 0 && i < 3);

    if (i == 0) {
      if (p.callDataSource().isEmpty()) {
        return nop();
      } else {
        return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
            .sourceId(hub.currentFrame().contextNumber())
            .targetId(hub.stamp() + 1)
            .sourceOffset(EWord.of(p.callDataSource().offset()))
            .size(p.callDataSource().length())
            .referenceSize(p.callDataSource().length())
            .phase(PHASE_SHA2_256_DATA)
            .setRipSha();
      }
    } else if (i == 1) {
      if (p.callDataSource().isEmpty()) {
        return new MmuCall(MMU_INST_MSTORE)
            .targetId(hub.stamp() + 1)
            .targetOffset(EWord.ZERO)
            .limb1(emptyHi)
            .limb2(emptyLo);
      } else {
        return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.stamp() + 1)
            .size(32)
            .phase(PHASE_SHA2_256_RESULT)
            .setRipSha();
      }
    } else {
      if (p.requestedReturnDataTarget().isEmpty()) {
        return nop();
      } else {
        return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.currentFrame().contextNumber())
            .sourceOffset(EWord.ZERO)
            .size(32)
            .referenceOffset(p.requestedReturnDataTarget().offset())
            .referenceSize(p.requestedReturnDataTarget().length());
      }
    }
  }

  public static MmuCall forSha2(final Hub hub, PrecompileInvocation p, int i) {
    return forRipeMd160Sha(
        hub,
        p,
        i,
        Bytes.fromHexString("e3b0c44298fc1c149afbf4c8996fb924"),
        Bytes.fromHexString("27ae41e4649b934ca495991b7852b855")); // SHA2-256({}) hi/lo
  }

  public static MmuCall forRipeMd160(final Hub hub, PrecompileInvocation p, int i) {
    return forRipeMd160Sha(
        hub,
        p,
        i,
        Bytes.fromHexString("9c1185a5"),
        Bytes.fromHexString("c5e9fc54612808977ee8f548b2258d31")); // RIPEMD160({}) hi/lo
  }

  public static MmuCall forIdentity(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 0 && i < 2);

    if (p.callDataSource().isEmpty()) {
      return nop();
    }

    if (i == 0) {
      return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .size(p.callDataSource().length())
          .referenceOffset(0)
          .referenceSize(p.callDataSource().length());
    } else {
      if (p.requestedReturnDataTarget().isEmpty()) {
        return nop();
      } else {
        return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.currentFrame().contextNumber())
            .sourceOffset(EWord.ZERO)
            .size(p.callDataSource().length())
            .referenceOffset(p.requestedReturnDataTarget().offset())
            .referenceSize(p.requestedReturnDataTarget().length());
      }
    }
  }

  public static MmuCall forEcAdd(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 0 && i < 3);
    if (i == 0) {
      return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .size(p.callDataSource().length())
          .referenceSize(128)
          .successBit(!p.ramFailure())
          .setEcData()
          .phase(PHASE_ECADD_DATA);
    } else if (i == 1) {
      return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.stamp() + 1)
          .size(64)
          .setEcData()
          .phase(PHASE_ECADD_RESULT);
    } else {
      return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.currentFrame().contextNumber())
          .targetOffset(EWord.of(p.requestedReturnDataTarget().offset()))
          .size(p.requestedReturnDataTarget().length())
          .referenceSize(64);
    }
  }

  public static MmuCall forEcMul(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 0 && i < 3);
    if (i == 0) {
      return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .size(p.callDataSource().length())
          .referenceSize(96)
          .successBit(!p.ramFailure())
          .setEcData()
          .phase(PHASE_ECMUL_DATA);
    } else if (i == 1) {
      return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.stamp() + 1)
          .size(64)
          .setEcData()
          .phase(PHASE_ECMUL_RESULT);
    } else {
      return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.currentFrame().contextNumber())
          .targetOffset(EWord.of(p.requestedReturnDataTarget().offset()))
          .size(p.requestedReturnDataTarget().length())
          .referenceSize(64);
    }
  }

  public static MmuCall forEcPairing(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 0 && i < 3);
    if (i == 0) {
      return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .size(p.callDataSource().length())
          .referenceSize(p.callDataSource().length())
          .successBit(!p.ramFailure())
          .setEcData()
          .phase(PHASE_PAIRING_DATA);
    } else if (i == 1) {
      if (p.callDataSource().isEmpty()) {
        return new MmuCall(MMU_INST_MSTORE).targetId(hub.stamp() + 1).limb2(Bytes.of(1));
      } else {
        return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.stamp() + 1)
            .size(32)
            .setEcData()
            .phase(PHASE_PAIRING_RESULT);
      }
    } else {
      return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.currentFrame().contextNumber())
          .targetOffset(EWord.of(p.requestedReturnDataTarget().offset()))
          .size(p.requestedReturnDataTarget().length())
          .referenceSize(32);
    }
  }

  public static MmuCall forBlake2f(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 0 && i < 4);
    if (i == 0) {
      return new MmuCall(MMU_INST_BLAKE)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset()))
          .successBit(!p.ramFailure())
          .limb1(EWord.of(((Blake2fMetadata) p.metadata()).r()))
          .limb2(EWord.of(((Blake2fMetadata) p.metadata()).f()))
          .setBlakeModexp()
          .phase(PHASE_BLAKE_PARAMETERS);
    } else if (i == 1) {
      return new MmuCall(MMU_INST_RAM_TO_EXO_WITH_PADDING)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(hub.stamp() + 1)
          .sourceOffset(EWord.of(p.callDataSource().offset() + 4))
          .size(208)
          .referenceSize(208)
          .setBlakeModexp()
          .phase(PHASE_BLAKE_DATA);
    } else if (i == 2) {
      return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(hub.stamp() + 1)
          .targetId(hub.stamp() + 1)
          .size(64)
          .setBlakeModexp()
          .phase(PHASE_BLAKE_RESULT);
    } else {
      if (p.requestedReturnDataTarget().isEmpty()) {
        return MmuCall.nop();
      } else {
        return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
            .sourceId(hub.stamp() + 1)
            .targetId(hub.currentFrame().contextNumber())
            .size(64)
            .referenceOffset(p.requestedReturnDataTarget().offset())
            .referenceSize(p.requestedReturnDataTarget().length());
      }
    }
  }

  public static MmuCall forModExp(final Hub hub, final PrecompileInvocation p, int i) {
    Preconditions.checkArgument(i >= 2 && i < 12);
    final ModExpMetadata m = (ModExpMetadata) p.metadata();
    final int prcContextNumber = hub.stamp() + 1;

    if (i == 2) {
      return new MmuCall(MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
          .sourceId(hub.currentFrame().contextNumber())
          .referenceOffset(p.callDataSource().offset())
          .referenceSize(p.callDataSource().length())
          .limb1(m.bbs().hi())
          .limb2(m.bbs().lo());
    } else if (i == 3) {
      return new MmuCall(MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
          .sourceId(hub.currentFrame().contextNumber())
          .sourceOffset(EWord.of(32))
          .referenceOffset(p.callDataSource().offset())
          .referenceSize(p.callDataSource().length())
          .limb1(m.ebs().hi())
          .limb2(m.ebs().lo());
    } else if (i == 4) {
      return new MmuCall(MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
          .sourceId(hub.currentFrame().contextNumber())
          .sourceOffset(EWord.of(64))
          .referenceOffset(p.callDataSource().offset())
          .referenceSize(p.callDataSource().length())
          .limb1(m.mbs().hi())
          .limb2(m.mbs().lo());
    } else if (i == 5) {
      return new MmuCall(MMU_INST_MLOAD)
          .sourceId(hub.currentFrame().contextNumber())
          .sourceOffset(EWord.of(p.callDataSource().offset() + 96 + m.bbs().toInt()))
          .limb1(m.rawLeadingWord().hi())
          .limb2(m.rawLeadingWord().lo());
    } else if (i == 7) {
      if (m.extractBase()) {
        return new MmuCall(MMU_INST_MODEXP_DATA)
            .sourceId(hub.currentFrame().contextNumber())
            .targetId(prcContextNumber)
            .sourceOffset(EWord.of(96))
            .size(m.bbs().toInt())
            .referenceOffset(p.callDataSource().offset())
            .referenceSize(p.callDataSource().length())
            .phase(PHASE_MODEXP_BASE)
            .setBlakeModexp();
      } else {
        return new MmuCall(MMU_INST_MODEXP_ZERO)
            .targetId(hub.stamp() + 1)
            .phase(PHASE_MODEXP_BASE)
            .setBlakeModexp();
      }
    } else if (i == 8) {
      if (m.extractExponent()) {
        return new MmuCall(MMU_INST_MODEXP_DATA)
            .sourceId(hub.currentFrame().contextNumber())
            .targetId(prcContextNumber)
            .sourceOffset(EWord.of(96 + m.bbs().toInt()))
            .size(m.ebs().toInt())
            .referenceOffset(p.callDataSource().offset())
            .referenceSize(p.callDataSource().length())
            .phase(PHASE_MODEXP_EXPONENT)
            .setBlakeModexp();
      } else {
        return new MmuCall(MMU_INST_MODEXP_ZERO)
            .targetId(hub.stamp() + 1)
            .phase(PHASE_MODEXP_EXPONENT)
            .setBlakeModexp();
      }
    } else if (i == 9) {
      return new MmuCall(MMU_INST_MODEXP_DATA)
          .sourceId(hub.currentFrame().contextNumber())
          .targetId(prcContextNumber)
          .sourceOffset(EWord.of(96 + m.bbs().toInt() + m.ebs().toInt()))
          .size(m.mbs().toInt())
          .referenceOffset(p.callDataSource().offset())
          .referenceSize(p.callDataSource().length())
          .phase(PHASE_MODEXP_MODULUS)
          .setBlakeModexp();
    } else if (i == 10) {
      return new MmuCall(MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(prcContextNumber)
          .targetId(prcContextNumber)
          .size(512)
          .phase(PHASE_MODEXP_RESULT)
          .setBlakeModexp();
    } else if (i == 11) {
      return new MmuCall(MMU_INST_RAM_TO_RAM_SANS_PADDING)
          .sourceId(prcContextNumber)
          .targetId(hub.currentFrame().contextNumber())
          .sourceOffset(EWord.of(512 - m.mbs().toInt()))
          .size(m.mbs().toInt())
          .referenceOffset(p.requestedReturnDataTarget().offset())
          .referenceSize(p.requestedReturnDataTarget().length());
    } else {
      throw new IllegalArgumentException("need a boolean");
    }
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscellaneousMmuFlag(this.enabled())
        .pMiscellaneousMmuInst(this.instruction())
        .pMiscellaneousMmuTgtId(this.sourceId())
        .pMiscellaneousMmuSrcId(this.targetId())
        .pMiscellaneousMmuAuxId(this.auxId())
        .pMiscellaneousMmuSrcOffsetHi(this.sourceOffset().hi())
        .pMiscellaneousMmuSrcOffsetLo(this.sourceOffset().lo())
        .pMiscellaneousMmuTgtOffsetLo(this.targetOffset().lo())
        .pMiscellaneousMmuSize(this.size())
        .pMiscellaneousMmuRefOffset(this.referenceOffset())
        .pMiscellaneousMmuRefSize(this.referenceSize())
        .pMiscellaneousMmuSuccessBit(this.successBit())
        .pMiscellaneousMmuLimb1(this.limb1())
        .pMiscellaneousMmuLimb2(this.limb2())
        .pMiscellaneousMmuExoSum(this.exoSum)
        .pMiscellaneousMmuPhase(this.phase());
  }
}
