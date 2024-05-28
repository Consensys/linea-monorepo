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

package net.consensys.linea.zktracer.module.ecdata;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECADD_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECADD_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECMUL_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECMUL_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECPAIRING_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECPAIRING_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECRECOVER_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECRECOVER_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.ECADD;
import static net.consensys.linea.zktracer.module.ecdata.Trace.ECMUL;
import static net.consensys.linea.zktracer.module.ecdata.Trace.ECPAIRING;
import static net.consensys.linea.zktracer.module.ecdata.Trace.ECRECOVER;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECADD_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECADD_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECMUL_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECMUL_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECPAIRING_DATA_MIN;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECPAIRING_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECRECOVER_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.INDEX_MAX_ECRECOVER_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.P_BN_HI;
import static net.consensys.linea.zktracer.module.ecdata.Trace.P_BN_LO;
import static net.consensys.linea.zktracer.module.ecdata.Trace.SECP256K1N_HI;
import static net.consensys.linea.zktracer.module.ecdata.Trace.SECP256K1N_LO;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECADD_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECADD_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECMUL_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECMUL_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECPAIRING_DATA_MIN;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECPAIRING_RESULT;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECRECOVER_DATA;
import static net.consensys.linea.zktracer.module.ecdata.Trace.TOTAL_SIZE_ECRECOVER_RESULT;
import static net.consensys.linea.zktracer.types.Containers.repeat;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.util.List;
import java.util.Optional;
import java.util.Set;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.Hash;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.crypto.SECPPublicKey;
import org.hyperledger.besu.crypto.SECPSignature;

@Accessors(fluent = true)
public class EcDataOperation extends ModuleOperation {
  private static final Set<Integer> EC_TYPES = Set.of(ECRECOVER, ECADD, ECMUL, ECPAIRING);
  private static final EWord P_BN = EWord.of(P_BN_HI, P_BN_LO);
  public static final EWord SECP256K1N = EWord.of(SECP256K1N_HI, SECP256K1N_LO);
  public static final int nBYTES_OF_DELTA_BYTES = 4; // TODO: from Corset ?

  private final Wcp wcp;
  private final Ext ext;

  @Getter private final long id;
  private final Bytes data;

  private final int ecType;
  private final int nRows;
  private final int nRowsData;
  private final int nRowsResult;

  @Getter private final List<Bytes> limb;
  private final List<Boolean> hurdle;
  @Getter private boolean internalChecksPassed;

  // WCP interaction
  private final List<Boolean> wcpFlag;
  private final List<Bytes> wcpArg1Hi;
  private final List<Bytes> wcpArg1Lo;
  private final List<Bytes> wcpArg2Hi;
  private final List<Bytes> wcpArg2Lo;
  private final List<Boolean> wcpRes;
  private final List<OpCode> wcpInst;

  // EXT interaction
  private final List<Boolean> extFlag;
  private final List<Bytes> extArg1Hi;
  private final List<Bytes> extArg1Lo;
  private final List<Bytes> extArg2Hi;
  private final List<Bytes> extArg2Lo;
  private final List<Bytes> extArg3Hi;
  private final List<Bytes> extArg3Lo;
  private final List<Bytes> extResHi;
  private final List<Bytes> extResLo;
  private final List<OpCode> extInst;

  private Bytes returnData;
  @Getter private boolean successBit;
  private boolean circuitSelectorEcrecover;

  // pairing-specific
  private final int totalPairings;

  private int getTotalSize(int ecType, boolean isData) {
    if (isData) {
      return switch (ecType) {
        case ECRECOVER -> TOTAL_SIZE_ECRECOVER_DATA;
        case ECADD -> TOTAL_SIZE_ECADD_DATA;
        case ECMUL -> TOTAL_SIZE_ECMUL_DATA;
        case ECPAIRING -> TOTAL_SIZE_ECPAIRING_DATA_MIN * this.totalPairings;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    } else {
      return switch (ecType) {
        case ECRECOVER -> successBit ? TOTAL_SIZE_ECRECOVER_RESULT : 0;
        case ECADD -> successBit ? TOTAL_SIZE_ECADD_RESULT : 0;
        case ECMUL -> successBit ? TOTAL_SIZE_ECMUL_RESULT : 0;
        case ECPAIRING -> successBit ? TOTAL_SIZE_ECPAIRING_RESULT : 0;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    }
  }

  private static short getPhase(int ecType, boolean isData) {
    if (isData) {
      return switch (ecType) {
        case ECRECOVER -> PHASE_ECRECOVER_DATA;
        case ECADD -> PHASE_ECADD_DATA;
        case ECMUL -> PHASE_ECMUL_DATA;
        case ECPAIRING -> PHASE_ECPAIRING_DATA;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    } else {
      return switch (ecType) {
        case ECRECOVER -> PHASE_ECRECOVER_RESULT;
        case ECADD -> PHASE_ECADD_RESULT;
        case ECMUL -> PHASE_ECMUL_RESULT;
        case ECPAIRING -> PHASE_ECPAIRING_RESULT;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    }
  }

  private int getIndexMax(int ecType, boolean isData) {
    if (isData) {
      return switch (ecType) {
        case ECRECOVER -> INDEX_MAX_ECRECOVER_DATA;
        case ECADD -> INDEX_MAX_ECADD_DATA;
        case ECMUL -> INDEX_MAX_ECMUL_DATA;
        case ECPAIRING -> (INDEX_MAX_ECPAIRING_DATA_MIN + 1) * this.totalPairings - 1;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    } else {
      return switch (ecType) {
        case ECRECOVER -> INDEX_MAX_ECRECOVER_RESULT;
        case ECADD -> INDEX_MAX_ECADD_RESULT;
        case ECMUL -> INDEX_MAX_ECMUL_RESULT;
        case ECPAIRING -> INDEX_MAX_ECPAIRING_RESULT;
        default -> throw new IllegalArgumentException("invalid EC type");
      };
    }
  }

  private EcDataOperation(Wcp wcp, Ext ext, int id, int ecType, Bytes data) {
    Preconditions.checkArgument(EC_TYPES.contains(ecType), "invalid EC type");

    final int minInputLength = ecType == ECMUL ? 96 : 128;
    if (data.size() < minInputLength) {
      this.data = leftPadTo(data, minInputLength);
    } else {
      this.data = data;
    }
    this.ecType = ecType;

    if (ecType == ECPAIRING) {
      this.totalPairings = data.size() / 192;
    } else {
      this.totalPairings = 0;
    }

    this.nRowsData = getIndexMax(ecType, true) + 1;
    this.nRowsResult = getIndexMax(ecType, false) + 1;
    this.nRows = this.nRowsData + this.nRowsResult;
    this.id = id;

    /*
    System.out.println(
        "(ecdataoperation filling time) previousId: "
            + Integer.toHexString(this.previousId)
            + " -> id: "
            + Integer.toHexString(this.id)
            + " , byteDelta: "
            + Arrays.stream(this.byteDelta).map(b -> Integer.toHexString(b.toInteger())).toList()
            + " , diff: "
            + Integer.toHexString(this.id - this.previousId - 1));

    System.out.println(
      "previousId: "
        + this.previousId
        + " -> id: "
        + this.id
        + " , byteDelta: "
        + Arrays.stream(this.byteDelta).map(b -> b.toInteger()).toList()
        + " , diff: " + (this.id - this.previousId - 1));
     */

    this.limb = repeat(Bytes.EMPTY, this.nRows);
    this.hurdle = repeat(false, this.nRows);

    this.wcpFlag = repeat(false, this.nRows);
    this.wcpArg1Hi = repeat(Bytes.EMPTY, this.nRows);
    this.wcpArg1Lo = repeat(Bytes.EMPTY, this.nRows);
    this.wcpArg2Hi = repeat(Bytes.EMPTY, this.nRows);
    this.wcpArg2Lo = repeat(Bytes.EMPTY, this.nRows);
    this.wcpRes = repeat(false, this.nRows);
    this.wcpInst = repeat(OpCode.INVALID, this.nRows);

    this.extFlag = repeat(false, this.nRows);
    this.extArg1Hi = repeat(Bytes.EMPTY, this.nRows);
    this.extArg1Lo = repeat(Bytes.EMPTY, this.nRows);
    this.extArg2Hi = repeat(Bytes.EMPTY, this.nRows);
    this.extArg2Lo = repeat(Bytes.EMPTY, this.nRows);
    this.extArg3Hi = repeat(Bytes.EMPTY, this.nRows);
    this.extArg3Lo = repeat(Bytes.EMPTY, this.nRows);
    this.extResHi = repeat(Bytes.EMPTY, this.nRows);
    this.extResLo = repeat(Bytes.EMPTY, this.nRows);
    this.extInst = repeat(OpCode.INVALID, this.nRows);

    this.wcp = wcp;
    this.ext = ext;

    switch (ecType) {
      case ECRECOVER -> handleRecover();
        // case ECADD -> handleAdd();
        // case ECMUL ->  handleMul();
        // case ECPAIRING -> handlePairing();
    }
  }

  public static EcDataOperation of(Wcp wcp, Ext ext, int id, final int ecType, Bytes data) {

    EcDataOperation ecDataRes = new EcDataOperation(wcp, ext, id, ecType, data);
    switch (ecType) {
      case ECRECOVER -> ecDataRes.handleRecover();
        // case ECADD -> ecDataRes.handleAdd();
        // case ECMUL -> ecDataRes.handleMul();
        // case ECPAIRING -> ecDataRes.handlePairing();
    }
    return ecDataRes;
  }

  private boolean callWcp(int i, OpCode wcpInst, EWord arg1, EWord arg2) {
    final boolean wcpRes =
        switch (wcpInst) {
          case LT -> this.wcp.callLT(arg1, arg2);
          case EQ -> this.wcp.callEQ(arg1, arg2);
          default -> throw new IllegalStateException("Unexpected value: " + wcpInst);
        };

    this.wcpFlag.set(i, true);
    this.wcpArg1Hi.set(i, arg1.hi());
    this.wcpArg1Lo.set(i, arg1.lo());
    this.wcpArg2Hi.set(i, arg2.hi());
    this.wcpArg2Lo.set(i, arg2.lo());
    this.wcpRes.set(i, wcpRes);
    this.wcpInst.set(i, wcpInst);
    return wcpRes;
  }

  private EWord callExt(int i, OpCode extInst, EWord arg1, EWord arg2, EWord arg3) {
    final EWord extRes = EWord.of(ext.call(extInst, arg1, arg2, arg3));

    this.extFlag.set(i, true);
    this.extArg1Hi.set(i, arg1.hi());
    this.extArg1Lo.set(i, arg1.lo());
    this.extArg2Hi.set(i, arg2.hi());
    this.extArg2Lo.set(i, arg2.lo());
    this.extArg3Hi.set(i, arg3.hi());
    this.extArg3Lo.set(i, arg3.lo());
    this.extResHi.set(i, extRes.hi());
    this.extResLo.set(i, extRes.lo());
    this.extInst.set(i, extInst);
    return extRes;
  }

  void handleRecover() {
    // Extract inputs
    final EWord h = EWord.of(this.data.slice(0, 32));
    final EWord v = EWord.of(this.data.slice(32, 32));
    final EWord r = EWord.of(this.data.slice(64, 32));
    final EWord s = EWord.of(this.data.slice(96, 32));

    // Set limb
    limb.set(0, h.hi());
    limb.set(1, h.lo());
    limb.set(2, v.hi());
    limb.set(3, v.lo());
    limb.set(4, r.hi());
    limb.set(5, r.lo());
    limb.set(6, s.hi());
    limb.set(7, s.lo());

    // Compute internal checks
    // row i
    boolean rIsInRange = callWcp(0, OpCode.LT, r, SECP256K1N); // r < secp256k1N

    // row i + 1
    boolean rIsPositive = callWcp(1, OpCode.LT, EWord.ZERO, r); // 0 < r

    // row i + 2
    boolean sIsInRange = callWcp(2, OpCode.LT, s, SECP256K1N); // s < secp256k1N

    // row i + 3
    boolean sIsPositive = callWcp(3, OpCode.LT, EWord.ZERO, s); // 0 < s

    // row i+ 4
    boolean vIs27 = callWcp(4, OpCode.EQ, v, EWord.of(27)); // v == 27

    // row i + 5
    boolean vIs28 = callWcp(5, OpCode.EQ, v, EWord.of(28)); // v == 28

    // Set hurdle
    hurdle.set(0, rIsInRange && rIsPositive);
    hurdle.set(1, sIsInRange && sIsPositive);
    hurdle.set(2, hurdle.get(0) && hurdle.get(1));
    hurdle.set(INDEX_MAX_ECRECOVER_DATA, hurdle.get(2) && (vIs27 || vIs28));

    // Set internal checks passed
    this.internalChecksPassed = hurdle.get(INDEX_MAX_ECRECOVER_DATA);

    EWord recoveredAddress = EWord.ZERO;

    // Compute recoveredAddress, successBit and set circuitSelectorEcrecover
    if (this.internalChecksPassed) {
      recoveredAddress = extractRecoveredAddress(h, v, r, s);
      this.circuitSelectorEcrecover = true;
    }

    successBit = !recoveredAddress.isZero();
    limb.set(8, recoveredAddress.hi());
    limb.set(9, recoveredAddress.lo());

    // Very unlikely edge case: if the ext module is never used elsewhere, we need to insert a
    // useless row, in order to trigger the construction of the first empty row, useful for the ext
    // lookup.
    // Because of the hashmap in the ext module, this useless row will only be inserted one time.
    // Tested by TestEcRecoverWithEmptyExt
    this.ext.callADDMOD(Bytes.EMPTY, Bytes.EMPTY, Bytes.EMPTY);
  }

  private static EWord extractRecoveredAddress(EWord h, EWord v, EWord r, EWord s) {
    SECP256K1 secp256K1 = new SECP256K1();
    try {
      Optional<SECPPublicKey> optionalRecoveredAddress =
          secp256K1.recoverPublicKeyFromSignature(
              h.toBytes(),
              SECPSignature.create(
                  r.toBigInteger(),
                  s.toBigInteger(),
                  (byte) (v.toInt() - 27),
                  SECP256K1N.toBigInteger()));
      return optionalRecoveredAddress
          .map(e -> EWord.of(Hash.keccak256(e.getEncodedBytes()).slice(32 - 20)))
          .orElse(EWord.ZERO);
    } catch (IllegalArgumentException e) {
      System.err.print(e);
      return EWord.ZERO;
    }
  }

  /*
  private void handlePointOnC1(final Bytes x, final Bytes y, int u, int i) {
    this.squares.set(
        6 * i + 2 * u, this.callExt(12 * i + 4 * u, OpCode.MULMOD, x, x, P_BN)); // x² mod p
    this.squares.set(
        1 + 2 * u + 6 * i, this.callExt(1 + 4 * u + 12 * i, OpCode.MULMOD, y, y, P_BN)); // y² mod p
    this.cubes.set(
        2 * u + 6 * i,
        this.callExt(
            2 + 4 * u + 12 * i,
            OpCode.MULMOD,
            this.squares.get(2 * u + 6 * i),
            x,
            P_BN)); // x³ mod p
    this.cubes.set(
        1 + 2 * u + 6 * i,
        this.callExt(
            3 + 4 * u + 12 * i,
            OpCode.ADDMOD,
            this.cubes.get(2 * u + 6 * i),
            Bytes.of(3),
            P_BN)); // x³ + 3 mod p

    this.comparisons.set(2 * u + 6 * i, this.callWcp(4 * u + 12 * i, OpCode.LT, x, P_BN)); // x < p
    this.comparisons.set(
        1 + 2 * u + 6 * i, this.callWcp(1 + 4 * u + 12 * i, OpCode.LT, y, P_BN)); // y < p

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

      this.comparisons.set(6 * i + 2, this.callWcp(12 * i + 3, OpCode.LT, aIm, P_BN));
      this.comparisons.set(6 * i + 3, this.callWcp(12 * i + 4, OpCode.LT, aRe, P_BN));
      this.comparisons.set(6 * i + 4, this.callWcp(12 * i + 5, OpCode.LT, bIm, P_BN));
      this.comparisons.set(6 * i + 5, this.callWcp(12 * i + 6, OpCode.LT, bRe, P_BN));
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
  */

  void trace(Trace trace, final int stamp, final long previousId) {
    final Bytes deltaByte =
        leftPadTo(Bytes.minimalBytes(id - previousId - 1), nBYTES_OF_DELTA_BYTES);

    for (int i = 0; i < this.nRows; i++) {
      boolean isData = i < this.nRowsData;
      trace
          .stamp(stamp)
          .id(id)
          .index(isData ? UnsignedByte.of(i) : UnsignedByte.of(i - this.nRowsData))
          .limb(limb.get(i))
          .totalSize(Bytes.ofUnsignedLong(getTotalSize(ecType, isData)))
          .phase(getPhase(ecType, isData))
          .indexMax(Bytes.ofUnsignedLong(getIndexMax(ecType, isData)))
          .successBit(successBit)
          .isEcrecoverData(ecType == ECRECOVER && isData)
          .isEcrecoverResult(ecType == ECRECOVER && !isData)
          .isEcaddData(ecType == ECADD && isData)
          .isEcaddResult(ecType == ECADD && !isData)
          .isEcmulData(ecType == ECMUL && isData)
          .isEcmulResult(ecType == ECMUL && !isData)
          .isEcpairingData(ecType == ECPAIRING && isData)
          .isEcpairingResult(ecType == ECPAIRING && !isData)
          .totalPairings(Bytes.ofUnsignedLong(totalPairings))
          .accPairings(ecType == ECPAIRING && isData ? Bytes.ofUnsignedLong(1 + i) : Bytes.of(0))
          .internalChecksPassed(internalChecksPassed)
          .hurdle(hurdle.get(i))
          .byteDelta(
              i < nBYTES_OF_DELTA_BYTES ? UnsignedByte.of(deltaByte.get(i)) : UnsignedByte.of(0))
          .circuitSelectorEcrecover(ecType == ECRECOVER && circuitSelectorEcrecover)
          .wcpFlag(wcpFlag.get(i))
          .wcpArg1Hi(wcpArg1Hi.get(i))
          .wcpArg1Lo(wcpArg1Lo.get(i))
          .wcpArg2Hi(wcpArg2Hi.get(i))
          .wcpArg2Lo(wcpArg2Lo.get(i))
          .wcpRes(wcpRes.get(i))
          .wcpInst(wcpInst.get(i).unsignedByteValue())
          .extFlag(extFlag.get(i))
          .extArg1Hi(extArg1Hi.get(i))
          .extArg1Lo(extArg1Lo.get(i))
          .extArg2Hi(extArg2Hi.get(i))
          .extArg2Lo(extArg2Lo.get(i))
          .extArg3Hi(extArg3Hi.get(i))
          .extArg3Lo(extArg3Lo.get(i))
          .extResHi(extResHi.get(i))
          .extResLo(extResLo.get(i))
          .extInst(extInst.get(i).unsignedByteValue())
          .fillAndValidateRow(); // TODO: add missing columns (stuff not related to ECRECOVER)
    }
  }

  @Override
  protected int computeLineCount() {
    return this.nRowsData + this.nRowsResult;
  }
}
