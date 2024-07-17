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

package net.consensys.linea.zktracer.module.limits.precompiles;

import static net.consensys.linea.zktracer.CurveOperations.*;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Stack;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.altbn128.AltBn128Fq2Point;
import org.hyperledger.besu.crypto.altbn128.AltBn128Point;
import org.hyperledger.besu.crypto.altbn128.Fq;
import org.hyperledger.besu.crypto.altbn128.Fq2;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
@RequiredArgsConstructor
@Accessors(fluent = true)
public final class EcPairingFinalExponentiations implements Module {
  private final Hub hub;
  @Getter private final Stack<EcPairingTallier> counts = new Stack<>();
  private static final int PRECOMPILE_BASE_GAS_FEE = 45_000; // cf EIP-1108
  private static final int PRECOMPILE_MILLER_LOOP_GAS_FEE = 34_000; // cf EIP-1108
  private static final int ECPAIRING_NB_BYTES_PER_MILLER_LOOP = 192;
  private static final int ECPAIRING_NB_BYTES_PER_SMALL_POINT = 64;
  private static final int ECPAIRING_NB_BYTES_PER_LARGE_POINT = 128;

  @Override
  public String moduleKey() {
    return "PRECOMPILE_ECPAIRING_FINAL_EXPONENTIATIONS";
  }

  @Override
  public void enterTransaction() {
    counts.push(new EcPairingTallier(0, 0, 0));
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  @Override
  public int lineCount() {
    int r = 0;
    for (EcPairingTallier count : this.counts) {
      r += (int) count.numberOfFinalExponentiations();
    }
    return r;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new UnsupportedOperationException("should never be called");
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    throw new UnsupportedOperationException("should never be called");
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {

    if (irrelevantOperation(frame)) {
      return;
    }

    final MemorySpan callDataSpan = hub.transients().op().callDataSegment();
    final long callDataSize = callDataSpan.length();
    final long callDataSizeMod192 = callDataSize % ECPAIRING_NB_BYTES_PER_MILLER_LOOP;
    final long nMillerLoop = (callDataSize / ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
    if (callDataSizeMod192 != 0) {
      log.warn(
          "[ECPAIRING] faulty call data size: {} ≡ {} mod {}",
          callDataSize,
          callDataSizeMod192,
          ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
      return;
    }

    final long gasAllowance = hub.transients().op().gasAllowanceForCall();
    final long precompileCost =
        PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * nMillerLoop;
    if (gasAllowance < precompileCost) {
      log.warn(
          "[ECPAIRING] insufficient gas: gas allowance = {}, precompile cost = {}",
          gasAllowance,
          precompileCost);
      return;
    }

    if (callDataSize == 0) {
      return;
    }

    /*
    At this point:
      - call data size is a positive multiple of 192
      - the precompile call is given sufficient gas
     */

    final EcPairingTallier currentEcpairingTallier = this.counts.pop();
    final long additionalRows = 12 * nMillerLoop + 2;

    if (!internalChecksPassed(frame, callDataSpan)) {
      this.counts.push(
          new EcPairingTallier(
              currentEcpairingTallier.numberOfMillerLoops(),
              currentEcpairingTallier.numberOfFinalExponentiations(),
              currentEcpairingTallier.numberOfG2MembershipTests()));
      return;
    }

    if (callDataContainsMalformedLargePoint(frame, callDataSpan)) {
      this.counts.push(
          new EcPairingTallier(
              currentEcpairingTallier.numberOfMillerLoops(),
              currentEcpairingTallier.numberOfFinalExponentiations(),
              currentEcpairingTallier.numberOfG2MembershipTests() + 1));
      return;
    }

    EcpairingCounts preciseCount = preciseCount(frame, callDataSpan);

    this.counts.push(
        new EcPairingTallier(
            currentEcpairingTallier.numberOfMillerLoops() + preciseCount.nontrivialPairs(),
            currentEcpairingTallier.numberOfFinalExponentiations() + preciseCount.nontrivialPairs()
                    == 0
                ? 0
                : 1,
            currentEcpairingTallier.numberOfG2MembershipTests() + preciseCount.membershipTests()));
  }

  public static boolean isHubFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (opCode.isCall()) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.ALTBN128_PAIRING)) {
        long length = hub.transients().op().callDataSegment().length();
        if (length % 192 != 0) {
          return true;
        }
        final long pairingCount = length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP;

        return hub.transients().op().gasAllowanceForCall()
            < PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * pairingCount;
      }
    }

    return false;
  }

  public static boolean isRamFailure(final Hub hub) {
    final MessageFrame frame = hub.messageFrame();
    long length = hub.transients().op().callDataSegment().length();

    if (length == 0) {
      return true;
    }

    for (int i = 0; i < length; i += 192) {
      final Bytes coordinates = frame.shadowReadMemory(i, 192);
      if (!isOnC1(coordinates.slice(0, 64)) || !isOnG2(coordinates.slice(64, 128))) {
        return true;
      }
    }

    return false;
  }

  public static long gasCost(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (irrelevantOperation(hub.messageFrame())) {
      return 0;
    }

    if (opCode.isCall()) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.ALTBN128_PAIRING)) {
        final long length = hub.transients().op().callDataSegment().length();
        final long nMillerLoop = (length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
        if (nMillerLoop * ECPAIRING_NB_BYTES_PER_MILLER_LOOP != length) {
          return 0;
        }

        return PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * nMillerLoop;
      }
    }

    return 0;
  }

  /**
   * Specifies if an opcode is irrelevant to the tracing of ECPAIRING.
   *
   * @param frame
   * @return true if the current operation is either not a call or would throw a stack exception;
   *     otherwise it returns true if the target of the CALL isn't the ECPAIRING precompile;
   */
  static boolean irrelevantOperation(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    if (!opCode.isCall() || frame.stackSize() < 2) {
      return true;
    }

    final Address target = Words.toAddress(frame.getStackItem(1));
    return !target.equals(Address.ALTBN128_PAIRING);
  }

  private boolean internalChecksPassed(MessageFrame frame, MemorySpan callData) {
    long offset = callData.offset();
    long nPairsOfPoints = callData.length() / ECPAIRING_NB_BYTES_PER_MILLER_LOOP;

    for (long i = 0; i < nPairsOfPoints; i++) {
      final Bytes coordinates = frame.shadowReadMemory(offset, ECPAIRING_NB_BYTES_PER_MILLER_LOOP);

      if (!isOnC1(coordinates.slice(0, ECPAIRING_NB_BYTES_PER_SMALL_POINT))) {
        return false;
      }

      final BigInteger pXIm =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT, 32));
      final BigInteger pXRe =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 32, 32));
      final BigInteger pYIm =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 64, 32));
      final BigInteger pYRe =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 96, 32));
      final Fq2 pX = Fq2.create(pXRe, pXIm);
      final Fq2 pY = Fq2.create(pYRe, pYIm);

      if (!pX.isValid() || !pY.isValid()) {
        return false;
      }

      offset += ECPAIRING_NB_BYTES_PER_MILLER_LOOP;
    }

    return true;
  }

  private boolean callDataContainsMalformedLargePoint(MessageFrame frame, MemorySpan callData) {
    long offset = callData.offset();
    long nPairsOfPoints = callData.length() / ECPAIRING_NB_BYTES_PER_MILLER_LOOP;

    for (long i = 0; i < nPairsOfPoints; i++) {
      final Bytes largeCoordinates =
          frame.shadowReadMemory(
              offset + ECPAIRING_NB_BYTES_PER_SMALL_POINT, ECPAIRING_NB_BYTES_PER_LARGE_POINT);

      // curve membership implicitly tested
      if (!isOnG2(largeCoordinates)) {
        return false;
      }

      offset += ECPAIRING_NB_BYTES_PER_MILLER_LOOP;
    }

    return true;
  }

  @Getter
  private class EcpairingCounts {
    int nontrivialPairs; // a pair of the form (A, B) with [A ≠ ∞] ∧ [B ≠ ∞]
    int membershipTests; // a pair of the form (A, B) with [A ≡ ∞] ∧ [B ≠ ∞]
    int trivialPairs; // a pair of the form (A, B) with  [B ≡ ∞]; likely useless ...

    private void incrementPairingPairs() {
      this.nontrivialPairs++;
    }

    private void incrementMembershipTests() {
      this.membershipTests++;
    }

    private void incrementTrivialPairs() {
      this.trivialPairs++;
    }
  }

  private EcpairingCounts preciseCount(MessageFrame frame, MemorySpan callData) {
    EcpairingCounts counts = new EcpairingCounts();

    long offset = callData.offset();
    long nPairsOfPoints = callData.length() / ECPAIRING_NB_BYTES_PER_MILLER_LOOP;

    for (long i = 0; i < nPairsOfPoints; i++) {
      final Bytes coordinates = frame.shadowReadMemory(offset, ECPAIRING_NB_BYTES_PER_MILLER_LOOP);

      final BigInteger smallPointX = extractParameter(coordinates.slice(0, 32));
      final BigInteger smallPointY = extractParameter(coordinates.slice(32, 32));
      final AltBn128Point smallPoint =
          new AltBn128Point(Fq.create(smallPointX), Fq.create(smallPointY));
      final boolean smallPointIsPointAtInfinity = smallPoint.isInfinity();

      final BigInteger largePointXIm =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT, 32));
      final BigInteger largePointXRe =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 32, 32));
      final BigInteger largePointYIm =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 64, 32));
      final BigInteger largePointYRe =
          extractParameter(coordinates.slice(ECPAIRING_NB_BYTES_PER_SMALL_POINT + 96, 32));
      final Fq2 largePointX = Fq2.create(largePointXRe, largePointXIm);
      final Fq2 largePointY = Fq2.create(largePointYRe, largePointYIm);
      AltBn128Fq2Point largePoint = new AltBn128Fq2Point(largePointX, largePointY);
      final boolean largePointIsPointAtInfinity = largePoint.isInfinity();

      if (largePointIsPointAtInfinity) {
        counts.incrementTrivialPairs();
      }

      if (!smallPointIsPointAtInfinity && !largePointIsPointAtInfinity) {
        counts.incrementPairingPairs();
      }

      if (smallPointIsPointAtInfinity && !largePointIsPointAtInfinity) {
        counts.incrementMembershipTests();
      }

      offset += ECPAIRING_NB_BYTES_PER_MILLER_LOOP;
    }

    return counts;
  }
}
