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

package net.consensys.linea.zktracer.module.exp;

import static java.lang.Math.max;
import static java.lang.Math.min;
import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import net.consensys.linea.zktracer.testing.EvmExtension;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@Slf4j
@ExtendWith(EvmExtension.class)
public class ExpTest {
  // Generates 128, 64, 2, 1 as LD
  private static final int[] LD_INDICES = new int[] {1, 2, 7, 8};
  private static final int[] C = new int[] {1, 2, 10, 15, 16, 17, 20, 31, 32};

  @Test
  void testExpLogSingleCase() {
    BytecodeCompiler program = BytecodeCompiler.newProgram().push(2).push(10).op(OpCode.EXP);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testModexpLogSingleCase() {
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            .push(1) // bbs
            .push(0)
            .op(OpCode.MSTORE)
            .push(1) // ebs
            .push(0x20)
            .op(OpCode.MSTORE)
            .push(1) // mbs
            .push(0x40)
            .op(OpCode.MSTORE)
            .push(
                Bytes.fromHexStringLenient(
                    "0x08090A0000000000000000000000000000000000000000000000000000000000")) // b, e,
            // m
            .push(0x60)
            .op(OpCode.MSTORE)
            .push(1) // retSize
            .push(0x9f) // retOffset
            .push(0x63) // argSize (cds)
            .push(0) // argOffset (cdo)
            .push(5) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testExpLogFFBlockCase() {
    for (int k = 0; k <= 32; k++) {
      Bytes exponent = Bytes.fromHexString(ffBlock(k));
      BytecodeCompiler program =
          BytecodeCompiler.newProgram().push(exponent).push(10).op(OpCode.EXP);
      BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
      bytecodeRunner.run();
    }
  }

  @Test
  void testExpLogFFAtCase() {
    for (int k = 1; k <= 32; k++) {
      Bytes exponent = Bytes.fromHexString(ffAt(k));
      BytecodeCompiler program =
          BytecodeCompiler.newProgram().push(exponent).push(10).op(OpCode.EXP);
      BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
      bytecodeRunner.run();
    }
  }

  @Disabled("EXP tests are disabled due to running for over 30 min.")
  @Test
  void testModexpLogFFBlockWithLDCase() {
    for (int ebsCutoff : C) {
      for (int cdsCutoff : C) {
        for (int k : C) {
          for (int LDIndex : LD_INDICES) {
            log.debug("k: " + k);
            log.debug("LDIndex: " + LDIndex);
            Bytes wordAfterBase = Bytes.fromHexStringLenient(ffBlockWithLd(k, LDIndex));
            BytecodeCompiler program =
                initProgramInvokingModexp(ebsCutoff, cdsCutoff, wordAfterBase);
            BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
            bytecodeRunner.run();
          }
        }
      }
    }
  }

  @Disabled("EXP tests are disabled due to running for over 30 min.")
  @Test
  void testModexpLogLDAtCase() {
    for (int ebsCutoff : C) {
      for (int cdsCutoff : C) {
        for (int k : C) {
          for (int ldIndex : LD_INDICES) {
            log.debug("k: " + k);
            log.debug("ldIndex: " + ldIndex);
            Bytes wordAfterBase = Bytes.fromHexStringLenient(ldAt(k, ldIndex));
            BytecodeCompiler program =
                initProgramInvokingModexp(ebsCutoff, cdsCutoff, wordAfterBase);
            BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
            bytecodeRunner.run();
          }
        }
      }
    }
  }

  @Test
  void testModexpLogFFBlockWithLDCaseSpecific() {
    final int ebsCutoff = 17;
    final int cdsCutoff = 17;
    final int k = 16;
    final int ldIndex = 1;

    Bytes wordAfterBase = Bytes.fromHexStringLenient(ffBlockWithLd(k, ldIndex));
    BytecodeCompiler program = initProgramInvokingModexp(ebsCutoff, cdsCutoff, wordAfterBase);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testModexpLogLDAtCaseSpecific() {
    final int ebsCutoff = 17;
    final int cdsCutoff = 17;
    final int k = 2;
    final int ldIndex = 1;

    Bytes wordAfterBase = Bytes.fromHexStringLenient(ldAt(k, ldIndex));
    BytecodeCompiler program = initProgramInvokingModexp(ebsCutoff, cdsCutoff, wordAfterBase);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Disabled("EXP tests are disabled due to running for over 30 min.")
  @Test
  void testModexpLogZerosCase() {
    for (int ebsCutoff : C) {
      for (int cdsCutoff : C) {
        Bytes wordAfterBase =
            Bytes.fromHexStringLenient(
                "0000000000000000000000000000000000000000000000000000000000000000");
        BytecodeCompiler program = initProgramInvokingModexp(ebsCutoff, cdsCutoff, wordAfterBase);
        BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
        bytecodeRunner.run();
      }
    }
  }

  private BytecodeCompiler initProgramInvokingModexp(
      int ebsCutoff, int cdsCutoff, Bytes wordAfterBase) {
    final int bbs = 0;
    final int minimalValidEbs = ebsCutoff;
    final int mbs = 0;
    final int minimalValidCds = cdsCutoff + 96 + bbs;

    return BytecodeCompiler.newProgram()
        .push(bbs) // bbs
        .push(0)
        .op(OpCode.MSTORE)
        .push(minimalValidEbs) // ebs
        .push(32)
        .op(OpCode.MSTORE)
        .push(mbs) // mbs
        .push(64)
        .op(OpCode.MSTORE)
        .push(wordAfterBase) // e
        .push(96 + bbs)
        .op(OpCode.MSTORE)
        .push(512) // retSize
        .push(minimalValidCds) // retOffset
        .push(minimalValidCds) // argSize (cds)
        .push(0) // argOffset (cdo)
        .push(5) // address
        .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
        .op(OpCode.STATICCALL);
  }

  @Disabled("EXP tests are disabled due to running for over 30 min.")
  @Test
  void testExpUtils() {
    // ExpLog case
    log.debug("FFBlock");
    for (int k = 0; k <= 32; k++) {
      log.debug(ffBlock(k));
    }

    log.debug("FFAt");
    for (int k = 1; k <= 32; k++) {
      log.debug(ffAt(k));
    }

    // ModexpLog case
    log.debug("FFBlockWithLD");
    for (int k : C) {
      for (int ldIndex : LD_INDICES) {
        log.debug(ffBlockWithLd(k, ldIndex));
      }
    }

    log.debug("LDAt");
    for (int k : C) {
      for (int ldIndex : LD_INDICES) {
        log.debug(ldAt(k, ldIndex));
      }
    }

    for (int ebsCutoff : C) {
      for (int cdsCutoff : C) {
        final int bbs = 0;
        final int minimalValidEbs = ebsCutoff;
        final int minimalValidCds = cdsCutoff + 96 + bbs;

        log.debug("minimalValidEbs: " + minimalValidEbs + ", minimalValidCds: " + minimalValidCds);
        log.debug("ebsCutoff: " + ebsCutoff + ", cdsCutoff: " + cdsCutoff);
        log.debug("###");

        assertThat(ebsCutoff).isEqualTo(min(minimalValidEbs, 32));
        assertThat(cdsCutoff).isEqualTo(min(max(minimalValidCds - (96 + bbs), 0), 32));
      }
    }
  }

  public static String ffBlock(int k) {
    if (k < 0 || k > 32) {
      throw new IllegalArgumentException("k must be between 0 and 32");
    }
    return "00".repeat(32 - k) + "ff".repeat(k);
  }

  public static String ffAt(int k) {
    if (k < 1 || k > 32) {
      throw new IllegalArgumentException("k must be between 1 and 32");
    }
    return "00".repeat(k - 1) + "ff" + "00".repeat(32 - k);
  }

  public static String ffBlockWithLd(int k, int LDIndex) {
    if (k < 1 || k > 32) {
      throw new IllegalArgumentException("k must be between 1 and 32");
    }
    if (LDIndex < 1 || LDIndex > 8) {
      throw new IllegalArgumentException("LDIndex must be between 1 and 8");
    }
    String ld =
        new BigInteger("0".repeat(LDIndex - 1) + "1" + "0".repeat(8 - LDIndex), 2).toString(16);
    if (k < 32) {
      return "00".repeat(32 - k - 1) + (ld.length() == 1 ? "0" + ld : ld) + "ff".repeat(k);
    }

    return "ff".repeat(k);
  }

  public static String ldAt(int k, int ldIndex) {
    if (k < 1 || k > 32) {
      throw new IllegalArgumentException("k must be between 1 and 32");
    }
    if (ldIndex < 1 || ldIndex > 8) {
      throw new IllegalArgumentException("ldindex must be between 1 and 8");
    }
    String ld =
        new BigInteger("0".repeat(ldIndex - 1) + "1" + "0".repeat(8 - ldIndex), 2).toString(16);

    return "00".repeat(k - 1) + (ld.length() == 1 ? "0" + ld : ld) + "00".repeat(32 - k);
  }
}
