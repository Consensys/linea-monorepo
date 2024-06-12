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

import java.util.Map;
import java.util.Set;

import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackOperation;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.commons.lang3.BooleanUtils;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.internal.Words;

@AllArgsConstructor(access = AccessLevel.PACKAGE)
class Type4PreComputation implements MmuPreComputation {
  private static final Set<Integer> TYPES =
      Set.of(MmuTrace.type4CC, MmuTrace.type4CD, MmuTrace.type4RD);

  @Override
  public MicroData dispatch(
      final MicroData microData,
      final OpCode opCode,
      final Map<Integer, StackOperation> stackOps,
      final CallStack callStack) {
    microData.opCode(opCode);
    microData.pointers(
        Pointers.builder()
            .stack1(stackOps.get(0).value().copy())
            .stack2(stackOps.get(1).value().copy())
            .build());

    Bytes value = stackOps.get(3).value().copy();
    microData.sizeImported(value.toInt());
    microData.value(value);

    setTern(microData, callStack);

    if (microData.skip()) {
      return microData;
    }

    int ternary = microData.ternary();

    if (ternary == MmuTrace.tern0) {
      dispatchTern0(microData, callStack);
    } else if (ternary == MmuTrace.tern1) {
      dispatchTern1(microData, callStack);
    } else if (ternary == MmuTrace.tern2) {
      dispatchTern2(microData);
    } else {
      microData.skip(true);
    }

    return microData;
  }

  @Override
  public MicroData preProcess(final MicroData microData, final CallStack callStack) {
    setContext(false, microData, callStack);

    return microData;
  }

  @Override
  public MicroData process(final MicroData microData, final CallStack callStack) {
    setContext(true, microData, callStack);
    int ternary = microData.ternary();

    if (ternary == MmuTrace.tern0) {
      updateTern0(microData);
    } else if (ternary == MmuTrace.tern1) {
      updateMicroOpType4Tern1(microData);
      updateOffsetType4Tern1(microData);
    } else if (ternary == MmuTrace.tern2) {
      updateMicroOpType4Tern2(microData);
      updateOffsetType4Tern2(microData);
    }

    return microData;
  }

  @Override
  public Set<Integer> preComputationTypes() {
    return TYPES;
  }

  private void updateOffsetType4Tern2(MicroData microData) {
    if (!microData.isFirstMicroInstruction()) {
      microData.targetLimbOffset().add(1);

      if (!microData.bits()[0]) {
        microData.targetByteOffset(UnsignedByte.ZERO);
      }
    }
  }

  private void updateMicroOpType4Tern2(MicroData microData) {
    boolean[] bits = microData.bits();
    UnsignedByte[] nibbles = microData.nibbles();

    if (bits[0]) {
      if (bits[2] && bits[3]) {
        microData.microOp(MmuTrace.KillingOne);
      } else {
        microData.size(microData.sizeImported());
        microData.microOp(MmuTrace.RamLimbExcision);
      }
    } else if (microData.isFirstMicroInstruction()) {
      if (bits[2]) {
        microData.microOp(MmuTrace.KillingOne);
      } else {
        microData.size(16 - nibbles[0].toInteger());
        microData.microOp(MmuTrace.RamLimbExcision);
      }
    } else if (microData.isLastPad()) {
      if (bits[3]) {
        microData.microOp(MmuTrace.KillingOne);
      } else {
        microData.size(nibbles[3].toInteger() + 1);
        microData.microOp(MmuTrace.RamLimbExcision);
      }
    } else {
      microData.microOp(MmuTrace.KillingOne);
    }
  }

  private void updateOffsetType4Tern1(MicroData microData) {
    if (microData.readPad().isRead(microData.processingRow())) {
      updateOffsetType4Tern1DataExtraction(microData);
    } else {
      updateOffsetType4Tern1ZeroPadding(microData);
    }
  }

  private void updateOffsetType4Tern1ZeroPadding(MicroData microData) {
    boolean[] bits = microData.bits();
    UnsignedByte[] nibbles = microData.nibbles();

    if (microData.isFirstPad()) {
      microData.targetByteOffset(nibbles[6]);
      if (bits[2]) {
        if (!bits[6]) {
          microData.targetLimbOffset().add(BooleanUtils.toInteger(bits[4]));
        } else {
          microData.targetLimbOffset().add(1);
        }
      } else if (bits[1]) {
        microData.targetLimbOffset().add(1);
      } else {
        microData.targetLimbOffset().add(BooleanUtils.toInteger(bits[4]));
      }
    }

    if (bits[3]) {
      microData.sourceLimbOffset(EWord.ZERO);
      microData.sourceByteOffset(UnsignedByte.ZERO);

      if (bits[4] && bits[5]) {
        // TODO
      } else {
        microData.size(1 + nibbles[7].toInteger() - nibbles[6].toInteger());
      }
    } else if (microData.isFirstPad()) {
      microData.targetByteOffset(nibbles[6]);
      microData.size(16 - nibbles[6].toInteger());
    } else {
      microData.targetLimbOffset().add(1);
      microData.targetByteOffset(UnsignedByte.ZERO);
      microData.size(1 + nibbles[7].toInteger());
    }
  }

  private void updateOffsetType4Tern1DataExtraction(MicroData microData) {
    boolean[] bits = microData.bits();
    UnsignedByte[] nibbles = microData.nibbles();
    int alignedInt = BooleanUtils.toInteger(microData.aligned());

    if (bits[2]) {
      microData.size(1 + nibbles[3].toInteger() - nibbles[2].toInteger());

      if (microData.processingRow() > 0) {
        microData.targetByteOffset(nibbles[6]);
        if (!bits[6]) {
          microData.targetLimbOffset().add(BooleanUtils.toInteger(bits[4]));
        } else {
          microData.targetLimbOffset().add(microData.targetLimbOffset()).add(1);
        }
      } else {
        if (microData.isFirstMicroInstruction()) {
          microData.size(15 - nibbles[2].toInteger() + 1);
        } else {
          microData.sourceLimbOffset().add(1);
        }

        int processingRow = microData.processingRow();
        if (processingRow == 1) {
          microData.targetLimbOffset().add(alignedInt + BooleanUtils.toInteger(bits[0]));
        }

        if (processingRow > 1) {
          microData.targetLimbOffset().add(1);
        }

        if (!microData.isFirstMicroInstruction() && microData.isRead()) {
          microData.sourceByteOffset(UnsignedByte.ZERO);
          microData.targetByteOffset(
              UnsignedByte.of(
                  nibbles[4].toInteger()
                      + 1
                      + (15 - nibbles[2].toInteger())
                      - 16 * (alignedInt + BooleanUtils.toInteger(bits[0]))));

          if (!microData.isLastRead()) {
            microData.size(16);
          } else {
            microData.size(nibbles[3].toInteger() + 1);
          }
        }
      }
    }
  }

  private void updateMicroOpType4Tern1(MicroData microData) {
    if (microData.readPad().isRead(microData.processingRow())) {
      updateMicroOpType4Tern1DataExtraction(microData);
    } else {
      updateMicroOpType4Tern1ZeroPadding(microData);
    }
  }

  private void updateMicroOpType4Tern1ZeroPadding(MicroData microData) {
    boolean[] bits = microData.bits();
    if (bits[3]) {
      if (bits[4] && bits[5]) {
        microData.microOp(MmuTrace.KillingOne);
      } else {
        microData.microOp(MmuTrace.RamLimbExcision);
      }
    } else if (!microData.isFirstPad()) {
      if (!microData.isLastPad()) {
        microData.microOp(MmuTrace.KillingOne);
      } else {
        if (!bits[5]) {
          microData.microOp(MmuTrace.RamLimbExcision);
        } else {
          microData.microOp(MmuTrace.KillingOne);
        }
      }
    } else if (bits[4]) {
      microData.microOp(MmuTrace.KillingOne);
    } else {
      microData.microOp(MmuTrace.RamLimbExcision);
    }
  }

  private void updateMicroOpType4Tern1DataExtraction(MicroData microData) {
    boolean[] bits = microData.bits();

    if (bits[2]) {
      if (!bits[6]) {
        modifyOneLimb(microData);
      } else {
        modifyTwoLimbs(microData);
      }
    } else {
      if (microData.isFirstMicroInstruction()) {
        if (!bits[0]) {
          modifyOneLimb(microData);
        } else {
          modifyTwoLimbs(microData);
        }
      } else {
        if (microData.remainingReads() != 0) {
          if (!microData.aligned()) {
            int preComputation = microData.precomputation();
            if (preComputation == MmuTrace.type4CC) {
              microData.microOp(MmuTrace.ExoToRam);
            } else if (preComputation == MmuTrace.type4CD) {
              if (!microData.info()) {
                microData.microOp(MmuTrace.RamToRam);
              } else {
                microData.microOp(MmuTrace.ExoToRam);
              }
            }
          } else {
            modifyTwoLimbs(microData);
          }
        } else {
          if (bits[1]) {
            modifyOneLimb(microData);
          } else {
            modifyTwoLimbs(microData);
          }
        }
      }
    }
  }

  private void updateTern0(MicroData microData) {
    if (microData.bits()[2]) {
      type4Tern0SingleMicroInstruction(microData);
    } else {
      type4Tern0MultipleMicroInstruction(microData);
    }
  }

  private void type4Tern0MultipleMicroInstruction(MicroData microData) {
    if (microData.processingRow() == 0) {
      if (microData.bits()[0]) {
        modifyTwoLimbs(microData);
      } else {
        modifyOneLimb(microData);
      }

      microData.size(16 - microData.nibbles()[2].toInteger());
    }
  }

  private void type4Tern0SingleMicroInstruction(final MicroData microData) {
    microData.size(microData.sizeImported());

    if (microData.bits()[3]) {
      modifyTwoLimbs(microData);
    } else {
      modifyOneLimb(microData);
    }
  }

  private void modifyOneLimb(MicroData microData) {
    int preComputation = microData.precomputation();
    if (preComputation == MmuTrace.type4CC) {
      microData.microOp(MmuTrace.ExoToRamSlideOverlappingChunk);
    } else if (preComputation == MmuTrace.type4RD) {
      microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
    } else if (preComputation == MmuTrace.type4CD) {
      if (!microData.info()) {
        microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
      } else {
        microData.microOp(MmuTrace.ExoToRamSlideOverlappingChunk);
      }
    } else {
      throw new IllegalArgumentException("Tern not supported");
    }
  }

  private void modifyTwoLimbs(MicroData microData) {
    int preComputation = microData.precomputation();
    if (preComputation == MmuTrace.type4CC) {
      microData.microOp(MmuTrace.ExoToRamSlideOverlappingChunk);
    } else if (preComputation == MmuTrace.type4RD) {
      microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
    } else if (preComputation == MmuTrace.type4CD) {
      if (!microData.info()) {
        microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
      } else {
        microData.microOp(MmuTrace.ExoToRamSlideOverlappingChunk);
      }
    } else {
      throw new IllegalArgumentException("Tern not supported");
    }
  }

  void setContext(final boolean isMicro, MicroData microData, final CallStack callStack) {
    microData.targetContext(callStack.current().contextNumber());

    switch (microData.opCode()) {
      case RETURNDATACOPY -> {
        microData.exoIsHash(false);
        microData.exoIsLog(false);
        microData.exoIsRom(false);
        microData.exoIsTxcd(false);
      }
      case CALLDATACOPY -> {
        microData.exoIsHash(false);
        microData.exoIsLog(false);
        microData.exoIsRom(false);

        if (isMicro && microData.isRead()) {
          microData.exoIsTxcd(microData.info());
        } else {
          microData.exoIsTxcd(false);
        }

        if (microData.callStackDepth() != 1) {
          microData.sourceContext(callStack.caller().contextNumber());
        } else {
          microData.sourceContext(0);
        }
      }
      case CODECOPY, EXTCODECOPY -> {
        microData.exoIsHash(false);
        microData.exoIsLog(false);

        microData.exoIsRom(isMicro && microData.isRead());
        microData.exoIsTxcd(false);

        microData.sourceContext(0);
      }
      default ->
          throw new UnsupportedOperationException(
              "OpCode.%s is not supported for MMU type 4 pre-processing and/or processing"
                  .formatted(microData.opCode()));
    }
  }

  void setTern(MicroData microData, final CallStack callStack) {
    final EWord off2 = EWord.of(microData.pointers().stack2());

    microData.referenceSize(calculateReferenceSize(microData, callStack));

    if ((microData.sizeImported() == 0 || microData.opCode() == OpCode.RETURNDATACOPY)
        && EWord.of(microData.referenceSize()).lessThan(off2.add(microData.sizeImported()))) {
      microData.skip(true);
      return;
    }

    EWord refSize = EWord.of(microData.referenceSize());
    EWord sizeImportedMinusOneEWord = EWord.of(microData.sizeImported() - 1);
    if (!off2.hi().isZero()) {
      microData.ternary(MmuTrace.tern2);
      microData.pointers().oob(true);
    } else if (off2.add(microData.sizeImported()).lessThan(refSize)) {
      microData.ternary(MmuTrace.tern0);
      microData.pointers().oob(false);

      EWord acc1 = refSize.subtract(off2).subtract(microData.sizeImported());
      microData.setAccsAtIndex(0, acc1);
    } else if (off2.lessThan(refSize)
        && refSize.lessOrEqualThan(sizeImportedMinusOneEWord.add(off2))) {
      microData.ternary(MmuTrace.tern1);
      microData.pointers().oob(false);

      EWord acc1 = EWord.of(microData.sizeImported() - 1 - microData.referenceSize()).add(off2);
      microData.setAccsAtIndex(0, acc1);

      EWord acc2 = EWord.of(microData.referenceSize() - 1).subtract(off2);
      microData.setAccsAtIndex(1, acc2);
    } else if (refSize.lessOrEqualThan(off2)) {
      microData.ternary(MmuTrace.tern2);
      microData.pointers().oob(true);

      EWord acc1 = EWord.ZERO.subtract(refSize);
      microData.setAccsAtIndex(0, acc1);
    }
  }

  private int calculateReferenceSize(final MicroData microData, final CallStack callStack) {
    OpCode opCode = microData.opCode();
    CallFrame topCallFrame = callStack.current();

    switch (opCode) {
      case CODECOPY -> topCallFrame.code().getSize();
      case EXTCODECOPY -> {
        final Address address = Words.toAddress(microData.value());
        topCallFrame.frame().getWorldUpdater().get(address).getCode().size();
      }
      case CALLDATACOPY -> callStack.caller().callDataRange().length();
      case RETURNDATACOPY -> topCallFrame.returnDataRange().length();
      default ->
          throw new IllegalArgumentException(
              "OpCode.%s not supported for type 4 reference size calculation.".formatted(opCode));
    }

    return 0;
  }

  private void dispatchTern0(final MicroData microData, final CallStack callStack) {
    Pointers pointers = microData.pointers();
    EWord off1Lo = EWord.of(pointers.stack1());
    EWord off2 = EWord.of(pointers.stack2());
    boolean[] bits = microData.bits();

    microData.referenceOffset(calculateReferenceOffset(microData, callStack));

    int refOffset = microData.referenceOffset();
    int sizeImported = microData.sizeImported();

    microData.setAccsAndNibblesAtIndex(2, EWord.of(refOffset).add(off2));
    microData.setAccsAndNibblesAtIndex(3, EWord.of(refOffset + sizeImported - 1).add(off2));
    microData.setAccsAndNibblesAtIndex(4, off1Lo);
    microData.setAccsAndNibblesAtIndex(5, EWord.of(sizeImported - 1).add(off1Lo));

    UnsignedByte[] nibbles = microData.nibbles();
    if (nibbles[4].toInteger() > nibbles[5].toInteger()) {
      bits[0] = true;
      nibbles[0] = UnsignedByte.of(nibbles[4].toInteger() - nibbles[2].toInteger() - 1);
    } else {
      bits[0] = false;
      nibbles[0] = UnsignedByte.of(nibbles[2].toInteger() - nibbles[4].toInteger());
    }

    if (nibbles[3].toInteger() > nibbles[5].toInteger()) {
      bits[1] = true;
      nibbles[1] = UnsignedByte.of(nibbles[3].toInteger() - nibbles[5].toInteger() - 1);
    } else {
      bits[1] = false;
      nibbles[1] = UnsignedByte.of(nibbles[5].toInteger() - nibbles[3].toInteger());
    }

    microData.readPad(
        ReadPad.builder()
            .totalNumberLimbs(
                microData.getAccsAtIndex(3).subtract(microData.getAccsAtIndex(2)).add(1).toInt())
            .totalNumberPaddingMicroInstructions(0)
            .build());

    if (nibbles[2].toInteger() == nibbles[4].toInteger()) {
      microData.aligned(true);
    }

    if (microData.readPad().totalNumber() == 1) {
      bits[2] = true;
      int sum = nibbles[4].toInteger() + sizeImported - 1;
      bits[3] = sum >= 16;
      nibbles[6] = UnsignedByte.of(sum % 16);
    }

    microData.offsets(
        Offsets.builder()
            .source(
                LimbByte.builder()
                    .limb(microData.getAccsAtIndex(2).copy())
                    .uByte(nibbles[2])
                    .build())
            .target(
                LimbByte.builder()
                    .limb(microData.getAccsAtIndex(4).copy())
                    .uByte(nibbles[4])
                    .build())
            .build());
  }

  private int calculateReferenceOffset(final MicroData microData, final CallStack callStack) {
    return switch (microData.opCode()) {
      case CODECOPY, EXTCODECOPY -> 0;
      case CALLDATACOPY -> callStack.caller().callDataRange().absolute().toInt();
      case RETURNDATACOPY -> callStack.current().returnDataRange().absolute().toInt();
      default ->
          throw new IllegalArgumentException(
              "OpCode.%s not supported for type 4 reference offset calculation"
                  .formatted(microData.opCode()));
    };
  }

  private void dispatchTern1(final MicroData microData, final CallStack callStack) {
    Pointers pointers = microData.pointers();
    EWord off1Lo = EWord.of(pointers.stack1());
    EWord off2 = EWord.of(pointers.stack2());
    boolean[] bits = microData.bits();

    microData.referenceOffset(calculateReferenceOffset(microData, callStack));

    int refOffset = microData.referenceOffset();
    int refSize = microData.referenceSize();
    int sizeImported = microData.sizeImported();

    microData.setAccsAndNibblesAtIndex(2, EWord.of(refOffset).add(off2));
    microData.setAccsAndNibblesAtIndex(3, EWord.of(refOffset + refSize - 1));
    microData.setAccsAndNibblesAtIndex(4, off1Lo);
    microData.setAccsAndNibblesAtIndex(5, EWord.of(refSize - 1).add(off1Lo).subtract(off2));
    microData.setAccsAndNibblesAtIndex(6, EWord.of(refSize).add(off1Lo).subtract(off2));
    microData.setAccsAndNibblesAtIndex(7, EWord.of(sizeImported - 1).add(off1Lo));

    UnsignedByte[] nibbles = microData.nibbles();

    if (nibbles[4].toInteger() > nibbles[2].toInteger()) {
      bits[0] = true;
      nibbles[0] = UnsignedByte.of(nibbles[4].toInteger() - nibbles[2].toInteger() - 1);
    } else {
      bits[0] = false;
      nibbles[0] = UnsignedByte.of(nibbles[2].toInteger() - nibbles[4].toInteger());
    }

    if (nibbles[3].toInteger() > nibbles[5].toInteger()) {
      bits[1] = true;
      nibbles[1] = UnsignedByte.of(nibbles[3].toInteger() - nibbles[5].toInteger() - 1);
    } else {
      bits[1] = false;
      nibbles[1] = UnsignedByte.of(nibbles[5].toInteger() - nibbles[3].toInteger());
    }

    microData.readPad(
        ReadPad.builder()
            .totalNumberLimbs(
                microData.getAccsAtIndex(3).subtract(microData.getAccsAtIndex(2)).add(1).toInt())
            .totalNumberPaddingMicroInstructions(
                microData.getAccsAtIndex(7).subtract(microData.getAccsAtIndex(6)).add(1).toInt())
            .build());

    microData.aligned(nibbles[2].toInteger() == nibbles[4].toInteger());

    ReadPad readPad = microData.readPad();
    bits[2] = readPad.totalNumberLimbs() == 1;
    bits[3] = readPad.totalNumberPaddingMicroInstructions() == 1;
    bits[4] = nibbles[5].toInteger() == 15;
    bits[5] = nibbles[7].toInteger() == 15;

    int calcNib = nibbles[4].toInteger() - nibbles[3].toInteger() - nibbles[2].toInteger();
    bits[6] = bits[2] && !(calcNib < 16);

    nibbles[8] = UnsignedByte.of(calcNib % 16);

    microData.sourceLimbOffset(microData.getAccsAtIndex(2).copy());
    microData.sourceByteOffset(nibbles[2]);

    microData.targetLimbOffset(microData.getAccsAtIndex(4).copy());
    microData.sourceByteOffset(nibbles[4]);
  }

  private void dispatchTern2(MicroData microData) {
    EWord off1 = EWord.of(microData.pointers().stack1());
    UnsignedByte[] nibbles = microData.nibbles();

    // ACC_3 & NIB_3
    EWord acc3 = off1.divide(16);
    microData.setAccsAtIndex(2, acc3);

    int nib3 = off1.mod(16).toInt();
    nibbles[2] = UnsignedByte.of(nib3);

    // ACC_4 & NIB_4
    EWord sum = EWord.of(microData.sizeImported() - 1).add(off1);
    EWord acc4 = sum.divide(16).add(off1);
    microData.setAccsAtIndex(3, acc4);

    int nib4 = sum.mod(16).toInt();
    nibbles[3] = UnsignedByte.of(nib4);

    microData.readPad(
        ReadPad.builder()
            .totalNumberLimbs(0)
            .totalNumberPaddingMicroInstructions(acc4.subtract(acc3).add(1).toInt())
            .build());

    boolean[] bits = microData.bits();

    bits[0] = microData.readPad().totalNumber() == 1;
    bits[2] = nib3 == 0;
    bits[3] = nib4 == 15;

    microData.targetLimbOffset(acc3);
    microData.targetByteOffset(UnsignedByte.of(nib3));
  }
}
