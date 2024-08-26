/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.shakiradata;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGEMO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_KECCAK_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_KECCAK_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_RIPEMD_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_RIPEMD_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_SHA2_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_SHA2_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WORD_SIZE;
import static net.consensys.linea.zktracer.module.shakiradata.HashType.KECCAK;
import static net.consensys.linea.zktracer.module.shakiradata.HashType.RIPEMD;
import static net.consensys.linea.zktracer.module.shakiradata.HashType.SHA256;
import static net.consensys.linea.zktracer.module.shakiradata.Trace.INDEX_MAX_RESULT;
import static net.consensys.linea.zktracer.types.Utils.rightPadTo;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.Hash;

@Accessors(fluent = true)
public class ShakiraDataOperation extends ModuleOperation {

  @Getter private final HashType hashType;
  private final Bytes hashInput;
  @Getter private final long ID;
  @Getter private final int inputSize;
  @Getter private final short lastNBytes;
  private final int indexMaxData;
  private Bytes32 result;

  // TODO: deprecate in favour of the version with the return data argument
  public ShakiraDataOperation(
      final long hubStamp, final HashType precompileType, final Bytes hashInput) {
    this.hashType = precompileType;
    this.ID = hubStamp + 1;
    this.hashInput = hashInput;
    this.inputSize = hashInput.size();
    this.lastNBytes = (short) (inputSize % LLARGE == 0 ? LLARGE : inputSize % LLARGE);
    // this.indexMaxData = Math.ceilDiv(inputSize, LLARGE) - 1;
    this.indexMaxData = (inputSize + LLARGEMO) / LLARGE - 1;
  }

  public ShakiraDataOperation(
      final int hubStamp,
      final HashType precompileType,
      final Bytes hashInput,
      final Bytes result) {
    this.hashType = precompileType;
    this.ID = hubStamp + 1;
    this.hashInput = hashInput;
    this.inputSize = hashInput.size();
    this.lastNBytes = (short) (inputSize % LLARGE == 0 ? LLARGE : inputSize % LLARGE);
    // this.indexMaxData = Math.ceilDiv(inputSize, LLARGE) - 1;
    this.indexMaxData = (inputSize + LLARGEMO) / LLARGE - 1;
    this.result = Bytes32.leftPad(result);
  }

  @Override
  protected int computeLineCount() {
    return indexMaxData + 1 + INDEX_MAX_RESULT + 1;
  }

  void trace(Trace trace, final int stamp) {

    this.result = Bytes32.leftPad(this.computeResult());
    traceData(trace, stamp);
    traceResult(trace, stamp);
  }

  private void traceData(Trace trace, final int stamp) {
    final boolean isShaData = hashType == SHA256;
    final boolean isKecData = hashType == KECCAK;
    final boolean isRipData = hashType == RIPEMD;
    final UnsignedByte phase =
        switch (hashType) {
          case SHA256 -> UnsignedByte.of(PHASE_SHA2_DATA);
          case KECCAK -> UnsignedByte.of(PHASE_KECCAK_DATA);
          case RIPEMD -> UnsignedByte.of(PHASE_RIPEMD_DATA);
        };

    for (int ct = 0; ct <= indexMaxData; ct++) {
      final boolean lastDataRow = ct == indexMaxData;
      trace
          .shakiraStamp(stamp)
          .id(ID)
          .phase(phase)
          .index(ct)
          .indexMax(indexMaxData)
          .limb(
              lastDataRow
                  ? rightPadTo(hashInput.slice(ct * LLARGE), LLARGE)
                  : hashInput.slice(ct * LLARGE, LLARGE))
          .nBytes(lastDataRow ? lastNBytes : LLARGE)
          .nBytesAcc(lastDataRow ? inputSize : (long) LLARGE * (ct + 1))
          .totalSize(inputSize)
          .isSha2Data(isShaData)
          .isKeccakData(isKecData)
          .isRipemdData(isRipData)
          .isSha2Result(false)
          .isKeccakResult(false)
          .isRipemdResult(false)
          .selectorKeccakResHi(false)
          .selectorSha2ResHi(false)
          .selectorRipemdResHi(false)
          .validateRow();
    }
  }

  private void traceResult(Trace trace, final int stamp) {
    final boolean isShaResult = hashType == SHA256;
    final boolean isKecResult = hashType == KECCAK;
    final boolean isRipResult = hashType == RIPEMD;
    final UnsignedByte phase =
        switch (hashType) {
          case SHA256 -> UnsignedByte.of(PHASE_SHA2_RESULT);
          case KECCAK -> UnsignedByte.of(PHASE_KECCAK_RESULT);
          case RIPEMD -> UnsignedByte.of(PHASE_RIPEMD_RESULT);
        };

    for (int ct = 0; ct <= INDEX_MAX_RESULT; ct++) {
      trace
          .shakiraStamp(stamp)
          .id(ID)
          .phase(phase)
          .index(ct)
          .indexMax(INDEX_MAX_RESULT)
          .isSha2Data(false)
          .isKeccakData(false)
          .isRipemdData(false)
          .isSha2Result(isShaResult)
          .isKeccakResult(isKecResult)
          .isRipemdResult(isRipResult)
          .nBytes((short) LLARGE)
          .totalSize(WORD_SIZE);

      switch (ct) {
        case 0 -> trace
            .limb(result.slice(0, LLARGE))
            .nBytesAcc(LLARGE)
            .selectorKeccakResHi(hashType == KECCAK)
            .selectorSha2ResHi(hashType == SHA256)
            .selectorRipemdResHi(hashType == RIPEMD)
            .validateRow();
        case 1 -> trace
            .limb(result.slice(LLARGE, LLARGE))
            .nBytesAcc(WORD_SIZE)
            .selectorKeccakResHi(false)
            .selectorSha2ResHi(false)
            .selectorRipemdResHi(false)
            .validateRow();
      }
    }
  }

  // TODO: this should die, we should use teh result from BESU
  private Bytes computeResult() {
    return switch (hashType) {
      case SHA256 -> Hash.sha256(hashInput);
      case KECCAK -> Hash.keccak256(hashInput);
      case RIPEMD -> Hash.ripemd160(hashInput);
    };
  }
}
