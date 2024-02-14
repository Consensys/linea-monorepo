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

package net.consensys.linea.zktracer.module.ec_data;

import static net.consensys.linea.zktracer.types.Containers.repeat;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.util.List;
import java.util.Set;

import com.google.common.base.Preconditions;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.CurveOperations;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@Accessors(fluent = true)
public class EcDataOperation extends ModuleOperation {
  private static final int EC_RECOVER = 1;
  private static final int EC_ADD = 6;
  private static final int EC_MUL = 7;
  private static final int EC_PAIRING = 8;
  private static final Set<Integer> EC_TYPES = Set.of(EC_RECOVER, EC_ADD, EC_MUL, EC_PAIRING);

  private static final EWord P =
      EWord.ofHexString("0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47");
  private static final EWord Q =
      EWord.ofHexString("0x30644E72E131A029B85045B68181585D2833E84879B9709143E1F593F0000001");
  private static final EWord SECP256_K1N =
      EWord.ofHexString("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f");

  private final Wcp wcp;
  private final Ext ext;

  private final int contextNumber;
  private final int contextNumberDelta;
  private final Bytes input;

  private final int ecType;
  private final int nRows;
  /** -1 if no switch off (i.e. preliminary checks passed) */
  private int hurdleSwitchOffRow;

  private final List<Boolean> comparisons;
  private final List<Boolean> equalities;
  private final List<Bytes> squares;
  private final List<Bytes> cubes;
  private final List<Bytes> limbs;

  // WCP interaction
  private final List<Bytes> wcpArg1;
  private final List<Bytes> wcpArg2;
  private final List<OpCode> wcpInst;
  private final List<Boolean> wcpRes;

  // EXT interaction
  private final List<Bytes> extArg1;
  private final List<Bytes> extArg2;
  private final List<Bytes> extArg3;
  private final List<OpCode> extInst;
  private final List<Bytes> extRes;

  // pairing-specific
  private int thisIsNotOnG2Index;
  private final int nPairings;

  private static int addressToEcType(final Address target) {
    if (target.equals(Address.ECREC)) {
      return EC_RECOVER;
    } else if (target.equals(Address.ALTBN128_ADD)) {
      return EC_ADD;
    } else if (target.equals(Address.ALTBN128_MUL)) {
      return EC_MUL;
    } else if (target.equals(Address.ALTBN128_PAIRING)) {
      return EC_PAIRING;
    } else {
      throw new IllegalArgumentException("invalid EC address");
    }
  }

  private static int ecTypeToNRows(int ecType, final Bytes input) {
    return switch (ecType) {
      case EC_RECOVER, EC_ADD -> 8;
      case EC_MUL -> 6;
      case EC_PAIRING -> input.size() / 16;
      default -> throw new IllegalArgumentException("invalid EC type");
    };
  }

  private EcDataOperation(
      Wcp wcp, Ext ext, int contextNumber, int previousContextNumber, int ecType, Bytes input) {
    Preconditions.checkArgument(EC_TYPES.contains(ecType), "invalid EC type");

    final int nRows = ecTypeToNRows(ecType, input);
    final int minInputLength = ecType == EC_MUL ? 96 : 128;
    if (input.size() < minInputLength) {
      this.input = leftPadTo(input, minInputLength);
    } else {
      this.input = input;
    }
    this.contextNumber = contextNumber;
    this.contextNumberDelta = contextNumber - previousContextNumber;
    this.ecType = ecType;
    this.nRows = nRows;
    this.thisIsNotOnG2Index = -1;
    if (ecType == EC_PAIRING) {
      this.nPairings = input.size() / 192;
    } else {
      this.nPairings = 0;
    }

    this.comparisons = repeat(false, nRows / 2);
    this.equalities = repeat(false, nRows);
    this.squares = repeat(Bytes.EMPTY, nRows / 2);
    this.cubes = repeat(Bytes.EMPTY, nRows / 2);
    this.limbs = repeat(Bytes.EMPTY, nRows / 2);

    this.wcpArg1 = repeat(Bytes.EMPTY, nRows);
    this.wcpArg2 = repeat(Bytes.EMPTY, nRows);
    this.wcpInst = repeat(OpCode.INVALID, nRows);
    this.wcpRes = repeat(false, nRows);

    this.extArg1 = repeat(Bytes.EMPTY, nRows);
    this.extArg2 = repeat(Bytes.EMPTY, nRows);
    this.extArg3 = repeat(Bytes.EMPTY, nRows);
    this.extInst = repeat(OpCode.INVALID, nRows);
    this.extRes = repeat(Bytes.EMPTY, nRows);

    this.wcp = wcp;
    this.ext = ext;
  }

  public static EcDataOperation of(
      Wcp wcp,
      Ext ext,
      final Address to,
      Bytes input,
      int currentContextNumber,
      int previousContextNumber) {
    final int ecType = addressToEcType(to);

    EcDataOperation r =
        new EcDataOperation(wcp, ext, currentContextNumber, previousContextNumber, ecType, input);
    switch (ecType) {
      case EC_RECOVER -> r.handleRecover();
      case EC_ADD -> r.handleAdd();
      case EC_MUL -> r.handleMul();
      case EC_PAIRING -> r.handlePairing();
    }
    return r;
  }

  private boolean preliminaryChecksPassed() {
    return this.hurdleSwitchOffRow == -1;
  }

  private boolean callWcp(int i, OpCode inst, Bytes arg1, Bytes arg2) {
    final boolean r =
        switch (inst) {
          case LT -> this.wcp.callLT(arg1, arg2);
          case EQ -> this.wcp.callEQ(arg1, arg2);
          default -> throw new IllegalStateException("Unexpected value: " + inst);
        };

    this.wcpArg1.set(i, arg1);
    this.wcpArg2.set(i, arg2);
    this.wcpInst.set(i, inst);
    this.wcpRes.set(i, r);
    return r;
  }

  private Bytes callExt(int i, OpCode opCode, Bytes arg1, Bytes arg2, Bytes arg3) {
    final Bytes result = ext.call(opCode, arg1, arg2, arg3);
    this.extArg1.set(i, arg1);
    this.extArg2.set(i, arg2);
    this.extArg3.set(i, arg3);
    this.extInst.set(i, opCode);
    this.extRes.set(i, result);

    return result;
  }

  private void fillHurdle() {
    for (int i = 0; i < this.nRows; i++) {
      boolean check = false;
      switch (i % 4) {
        case 1 -> {
          check = true;
        }
        case 0, 2 -> {
          check = this.comparisons.get(i / 2);
        }
        case 3 -> {
          check = this.equalities.get(i);
        }
      }
      if (!check) {
        this.hurdleSwitchOffRow = i;
        return;
      }
    }

    this.hurdleSwitchOffRow = -1;
  }

  void handleRecover() {
    final Bytes v = this.input.slice(32, 32);
    final Bytes r = this.input.slice(64, 32);
    final Bytes s = this.input.slice(96, 32);

    this.comparisons.set(0, this.callWcp(0, OpCode.LT, r, SECP256_K1N)); // r < secp256k1N
    this.comparisons.set(2, this.callWcp(1, OpCode.LT, s, SECP256_K1N)); // s < secp256k1N
    this.comparisons.set(1, this.callWcp(2, OpCode.LT, Bytes.EMPTY, r)); // 0 < r
    this.comparisons.set(3, this.callWcp(3, OpCode.LT, Bytes.EMPTY, s)); // 0 < s
    this.equalities.set(1, this.callWcp(4, OpCode.EQ, v, Bytes.of(27))); // v == 27
    this.equalities.set(2, this.callWcp(5, OpCode.EQ, v, Bytes.of(28))); // v == 28
    this.equalities.set(3, this.equalities.get(1) || this.equalities.get(2));
    this.equalities.set(7, true);

    this.fillHurdle();

    // Very unlikely edge case: if the ext module is never used elsewhere, we need to insert a
    // useless row, in order to trigger the construction of the first empty row, useful for the ext
    // lookup.
    // Because of the hashmap in the ext module, this useless row will only be inserted one time.
    // Tested by TestEcRecoverWithEmptyExt
    this.ext.callADDMOD(Bytes.EMPTY, Bytes.EMPTY, Bytes.EMPTY);
  }

  private void handlePointOnC1(final Bytes x, final Bytes y, int u, int i) {
    this.squares.set(
        6 * i + 2 * u, this.callExt(12 * i + 4 * u, OpCode.MULMOD, x, x, P)); // x² mod p
    this.squares.set(
        1 + 2 * u + 6 * i, this.callExt(1 + 4 * u + 12 * i, OpCode.MULMOD, y, y, P)); // y² mod p
    this.cubes.set(
        2 * u + 6 * i,
        this.callExt(
            2 + 4 * u + 12 * i, OpCode.MULMOD, this.squares.get(2 * u + 6 * i), x, P)); // x³ mod p
    this.cubes.set(
        1 + 2 * u + 6 * i,
        this.callExt(
            3 + 4 * u + 12 * i,
            OpCode.ADDMOD,
            this.cubes.get(2 * u + 6 * i),
            Bytes.of(3),
            P)); // x³ + 3 mod p

    this.comparisons.set(2 * u + 6 * i, this.callWcp(4 * u + 12 * i, OpCode.LT, x, P)); // x < p
    this.comparisons.set(
        1 + 2 * u + 6 * i, this.callWcp(1 + 4 * u + 12 * i, OpCode.LT, y, P)); // y < p

    this.equalities.set(
        1 + 4 * u + 12 * i,
        this.callWcp(
            2 + 4 * u + 12 * i,
            OpCode.EQ,
            this.squares.get(1 + 2 * u + 6 * i),
            this.cubes.get(1 + 2 * u + 6 * i))); // y² = x³ + 3
    this.equalities.set(2 + 4 * u + 12 * i, x.isZero() && y.isZero()); // (x, y) == (0, 0)
    this.equalities.set(
        3 + 4 * u + 12 * i,
        this.equalities.get(1 + 4 * u + 12 * i) || this.equalities.get(2 + 4 * u + 12 * i));
  }

  void handleAdd() {
    for (int u = 0; u < 2; u++) {
      final Bytes x = this.input.slice(64 * u, 32);
      final Bytes y = this.input.slice(64 * u + 32, 32);
      this.handlePointOnC1(x, y, u, 0);
    }
  }

  void handleMul() {
    final Bytes x = this.input.slice(0, 32);
    final Bytes y = this.input.slice(32, 32);
    this.handlePointOnC1(x, y, 0, 0);
    this.comparisons.set(2, true);
    this.fillHurdle();
  }

  void handlePairing() {
    for (int i = 0; i < this.nPairings; i++) {
      final Bytes x = this.input.slice(i * 192, 32);
      final Bytes y = this.input.slice(i * 192 + 32, 32);
      final Bytes aIm = this.input.slice(i * 192 + 64, 32);
      final Bytes aRe = this.input.slice(i * 192 + 96, 32);
      final Bytes bIm = this.input.slice(i * 192 + 128, 32);
      final Bytes bRe = this.input.slice(i * 192 + 160, 32);

      this.handlePointOnC1(x, y, 0, i);

      this.comparisons.set(6 * i + 2, this.callWcp(12 * i + 3, OpCode.LT, aIm, P));
      this.comparisons.set(6 * i + 3, this.callWcp(12 * i + 4, OpCode.LT, aRe, P));
      this.comparisons.set(6 * i + 4, this.callWcp(12 * i + 5, OpCode.LT, bIm, P));
      this.comparisons.set(6 * i + 5, this.callWcp(12 * i + 6, OpCode.LT, bRe, P));
      this.equalities.set(12 * i + 7, true);
      this.equalities.set(12 * i + 11, true);
    }

    this.fillHurdle();

    if (this.preliminaryChecksPassed()) {
      for (int i = 0; i < this.nPairings; i++) {
        if (!CurveOperations.isOnG2(this.input.slice(192 * i + 64, 192 - 64))) {
          this.thisIsNotOnG2Index = i;
          break;
        }
      }
    }
  }

  private void traceRow(Trace trace, int i) {
    trace.stamp(Bytes.ofUnsignedShort(this.contextNumber));
    // TODO: the rest
  }

  void trace(Trace trace) {
    for (int i = 0; i < this.nRows; i++) {
      this.traceRow(trace, i);
    }
  }

  @Override
  protected int computeLineCount() {
    return this.nRows;
  }
}
