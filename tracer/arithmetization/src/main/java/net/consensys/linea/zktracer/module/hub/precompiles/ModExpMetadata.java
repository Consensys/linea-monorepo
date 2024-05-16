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

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;

public record ModExpMetadata(
    boolean extractBbs,
    boolean extractEbs,
    boolean extractMbs,
    EWord bbs,
    EWord ebs,
    EWord mbs,
    boolean loadRawLeadingWord,
    EWord rawLeadingWord,
    boolean extractModulus,
    boolean extractBase,
    boolean extractExponent)
    implements PrecompileMetadata {
  public static ModExpMetadata of(final Hub hub) {
    final MemorySpan callDataSource = hub.transients().op().callDataSegment();
    final boolean extractBbs = !callDataSource.isEmpty();
    final boolean extractEbs = callDataSource.length() > 32;
    final boolean extractMbs = callDataSource.length() > 64;

    final int bbsShift = 32 - (int) Math.min(32, callDataSource.length());
    final Bytes rawBbs =
        extractBbs ? hub.messageFrame().shadowReadMemory(callDataSource.offset(), 32) : Bytes.EMPTY;
    final EWord bbs = EWord.of(rawBbs.shiftRight(bbsShift).shiftLeft(bbsShift));

    final int ebsShift = extractEbs ? 32 - (int) Math.min(32, callDataSource.length() - 32) : 0;
    final Bytes rawEbs =
        extractEbs
            ? hub.messageFrame().shadowReadMemory(callDataSource.offset() + 32, 32)
            : Bytes.EMPTY;
    final EWord ebs = EWord.of(rawEbs.shiftRight(ebsShift).shiftLeft(ebsShift));

    final int mbsShift = extractMbs ? 32 - (int) Math.min(32, callDataSource.length() - 64) : 0;
    final Bytes rawMbs =
        extractMbs
            ? hub.messageFrame().shadowReadMemory(callDataSource.offset() + 64, 32)
            : Bytes.EMPTY;
    final EWord mbs = EWord.of(rawMbs.shiftRight(mbsShift).shiftLeft(mbsShift));

    // TODO: maybe do not use intValueExact() here and just convert to int
    // TODO: checks over size may be done later
    final int bbsInt = bbs.toUnsignedBigInteger().intValueExact();
    final int ebsInt = ebs.toUnsignedBigInteger().intValueExact();

    final boolean loadRawLeadingWord = callDataSource.length() > 96 + bbsInt && !ebs.isZero();

    final EWord rawLeadingWord =
        loadRawLeadingWord
            ? EWord.of(
                hub.messageFrame().shadowReadMemory(callDataSource.offset() + 96 + bbsInt, 32))
            : EWord.ZERO;

    final boolean extractModulus =
        (callDataSource.length() > 96 + bbsInt + ebsInt) && !mbs.isZero();
    final boolean extractBase = extractModulus && !bbs.isZero();
    final boolean extractExponent = extractModulus && !ebs.isZero();

    return new ModExpMetadata(
        extractBbs,
        extractEbs,
        extractMbs,
        bbs,
        ebs,
        mbs,
        loadRawLeadingWord,
        rawLeadingWord,
        extractModulus,
        extractBase,
        extractExponent);
  }
}
