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

package net.consensys.linea.zktracer.module.rlptxn.cancun.phaseSection;

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.rlpUtils.RlpUtils.BYTES_PREFIX_SHORT_INT;
import static net.consensys.linea.zktracer.module.rlputilsOld.Pattern.outerRlpSize;
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.rlpUtils.InstructionByteStringPrefix;
import net.consensys.linea.zktracer.module.rlpUtils.InstructionBytes32;
import net.consensys.linea.zktracer.module.rlpUtils.RlpUtils;
import net.consensys.linea.zktracer.module.rlptxn.cancun.GenericTracedValue;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;

public class AccessListPhaseSection extends PhaseSection {
  private int phaseSize;
  private int totalAddress;
  private int totalKeys;
  private final InstructionByteStringPrefix accessListRlpPrefix;
  private final List<EntrySubSection> entries;

  public AccessListPhaseSection(RlpUtils rlpUtils, TransactionProcessingMetadata tx) {
    final List<AccessListEntry> accessList = tx.getBesuTransaction().getAccessList().get();
    final List<Short> rlpKeysSizeList = new ArrayList<>(accessList.size());
    for (AccessListEntry entry : accessList) {
      rlpKeysSizeList.add((short) outerRlpSize(33 * entry.storageKeys().size()));
    }
    final List<Short> entrySizeList = new ArrayList<>(accessList.size());
    for (Short rlpKeySize : rlpKeysSizeList) {
      entrySizeList.add((short) outerRlpSize(21 + rlpKeySize));
    }
    phaseSize = entrySizeList.stream().mapToInt(entry -> entry).sum();

    final InstructionByteStringPrefix accessListRlpPrefixCall =
        new InstructionByteStringPrefix(phaseSize, (byte) 0x00, true);
    accessListRlpPrefix = (InstructionByteStringPrefix) rlpUtils.call(accessListRlpPrefixCall);

    entries = new ArrayList<>(accessList.size());
    for (int i = 0; i < accessList.size(); i++) {
      entries.add(
          new EntrySubSection(
              rlpUtils, accessList.get(i), entrySizeList.get(i), rlpKeysSizeList.get(i)));
    }
  }

  @Override
  protected void traceComputationsRows(
      Trace.Rlptxn trace, TransactionProcessingMetadata tx, GenericTracedValue tracedValues) {
    totalAddress = tx.numberOfWarmedAddresses();
    totalKeys = tx.numberOfWarmedStorageKeys();

    // Phase RlpPrefix
    tracePreValues(trace, tracedValues);
    accessListRlpPrefix.traceRlpTxn(trace, tracedValues, true, true, true, 0);
    phaseSize -= accessListRlpPrefix.rlpPrefixByteSize();
    trace.pCmpIsPrefix(true).pCmpTmp1(phaseSize).pCmpTmp2(totalAddress).pCmpTmp3(totalKeys);
    trace.phaseEnd(entries.isEmpty());
    tracePostValues(trace, tracedValues);

    // trace each entry
    for (EntrySubSection entry : entries) {
      entry.trace(trace, tracedValues);
    }
  }

  @Override
  protected void traceIsPhaseX(Trace.Rlptxn trace) {
    trace.isAccessList(true);
  }

  @Override
  public int lineCount() {
    return 1 // 1 for transaction row
        + 1 // + 1 for global Rlp prefix
        + entries.stream().mapToInt(EntrySubSection::lineCount).sum(); // + the entries
  }

  private class EntrySubSection {
    private final InstructionByteStringPrefix entryRlpPrefix;
    private final Address address;
    private final InstructionByteStringPrefix keysRlpPrefix;
    private final List<InstructionBytes32> keys;

    private EntrySubSection(
        RlpUtils rlpUtils, AccessListEntry entry, short entryRlpSize, short rlpKeysSize) {
      final InstructionByteStringPrefix entryRlpPrefixCall =
          new InstructionByteStringPrefix(entryRlpSize, (byte) 0x00, true);
      entryRlpPrefix = (InstructionByteStringPrefix) rlpUtils.call(entryRlpPrefixCall);

      address = entry.address();

      final InstructionByteStringPrefix keysRlpPrefixCall =
          new InstructionByteStringPrefix(rlpKeysSize, (byte) 0x00, true);
      keysRlpPrefix = (InstructionByteStringPrefix) rlpUtils.call(keysRlpPrefixCall);

      keys = new ArrayList<>(entry.storageKeys().size());
      for (Bytes32 key : entry.storageKeys()) {
        final InstructionBytes32 call = new InstructionBytes32(key);
        keys.add((InstructionBytes32) rlpUtils.call(call));
      }
    }

    private int lineCount() {
      return 1 // 1 for entry RlpPrefix
          + 2 // 2 for the Address
          + 1 // 1 for the RlpPrefix of the
          // list of keys
          + 3 * keys.size(); // 3 per keys
    }

    public void trace(Trace.Rlptxn trace, GenericTracedValue tracedValues) {
      int tupleSize = entryRlpPrefix.rlpPrefixByteSize();
      int totalStorageForThisAddress = keys.size();

      // trace entry RlpPrefix
      tracePreValues(trace, tracedValues);
      entryRlpPrefix.traceRlpTxn(trace, tracedValues, true, true, true, 0);
      phaseSize -= tupleSize;
      trace
          .pCmpIsPrefix(true)
          .pCmpIsAddress(true)
          .pCmpTmp1(phaseSize)
          .pCmpTmp2(totalAddress)
          .pCmpTmp3(totalKeys)
          .pCmpTmp4(tupleSize)
          .pCmpTmp5(highPart(address))
          .pCmpTmp6(lowPart(address))
          .pCmpTmp7(totalStorageForThisAddress);
      tracePostValues(trace, tracedValues);

      // trace RLP(address)
      // first Limb
      tracePreValues(trace, tracedValues);
      trace
          .ctMax(1)
          .limbConstructed(true)
          .lt(true)
          .lx(true)
          .limb(Bytes16.rightPad(Bytes.concatenate(BYTES_PREFIX_SHORT_INT, address.slice(0, 4))))
          .nBytes(5);
      phaseSize -= 5;
      tupleSize -= 5;
      tracedValues.decrementLtAndLxSizeBy(5);
      trace
          .pCmpIsAddress(true)
          .pCmpTmp1(phaseSize)
          .pCmpTmp2(totalAddress)
          .pCmpTmp3(totalKeys)
          .pCmpTmp4(tupleSize)
          .pCmpTmp5(highPart(address))
          .pCmpTmp6(lowPart(address))
          .pCmpTmp7(totalStorageForThisAddress);
      tracePostValues(trace, tracedValues);

      // second limb
      tracePreValues(trace, tracedValues);
      trace
          .ct(1)
          .ctMax(1)
          .limbConstructed(true)
          .lt(true)
          .lx(true)
          .limb(address.slice(4, LLARGE))
          .nBytes(LLARGE);
      totalAddress -= 1;
      phaseSize -= LLARGE;
      tupleSize -= LLARGE;
      tracedValues.decrementLtAndLxSizeBy(LLARGE);
      trace
          .pCmpIsAddress(true)
          .pCmpTmp1(phaseSize)
          .pCmpTmp2(totalAddress)
          .pCmpTmp3(totalKeys)
          .pCmpTmp4(tupleSize)
          .pCmpTmp5(highPart(address))
          .pCmpTmp6(lowPart(address))
          .pCmpTmp7(totalStorageForThisAddress);
      tracePostValues(trace, tracedValues);

      // trace RLP prefix (keys)
      tracePreValues(trace, tracedValues);
      keysRlpPrefix.traceRlpTxn(trace, tracedValues, true, true, true, 0);
      phaseSize -= keysRlpPrefix.rlpPrefixByteSize();
      tupleSize -= keysRlpPrefix.rlpPrefixByteSize();
      trace
          .pCmpIsPrefix(true)
          .pCmpIsStorage(true)
          .pCmpTmp1(phaseSize)
          .pCmpTmp2(totalAddress)
          .pCmpTmp3(totalKeys)
          .pCmpTmp4(tupleSize)
          .pCmpTmp5(highPart(address))
          .pCmpTmp6(lowPart(address))
          .pCmpTmp7(totalStorageForThisAddress);
      trace.phaseEnd(phaseSize == 0);
      tracePostValues(trace, tracedValues);

      // optionally trace RLP(keys)
      for (InstructionBytes32 key : keys) {
        // rlp Prefix
        tracePreValues(trace, tracedValues);
        key.traceRlpTxn(trace, tracedValues, true, true, true, 0);
        phaseSize -= 1;
        tupleSize -= 1;
        trace
            .pCmpIsStorage(true)
            .pCmpTmp1(phaseSize)
            .pCmpTmp2(totalAddress)
            .pCmpTmp3(totalKeys)
            .pCmpTmp4(tupleSize)
            .pCmpTmp5(highPart(address))
            .pCmpTmp6(lowPart(address))
            .pCmpTmp7(totalStorageForThisAddress);
        tracePostValues(trace, tracedValues);

        // key hi
        tracePreValues(trace, tracedValues);
        key.traceRlpTxn(trace, tracedValues, true, true, true, 1);
        phaseSize -= LLARGE;
        tupleSize -= LLARGE;
        trace
            .pCmpIsStorage(true)
            .pCmpTmp1(phaseSize)
            .pCmpTmp2(totalAddress)
            .pCmpTmp3(totalKeys)
            .pCmpTmp4(tupleSize)
            .pCmpTmp5(highPart(address))
            .pCmpTmp6(lowPart(address))
            .pCmpTmp7(totalStorageForThisAddress);
        tracePostValues(trace, tracedValues);

        // key lo
        tracePreValues(trace, tracedValues);
        key.traceRlpTxn(trace, tracedValues, true, true, true, 2);
        phaseSize -= LLARGE;
        tupleSize -= LLARGE;
        totalKeys -= 1;
        totalStorageForThisAddress -= 1;
        trace
            .pCmpIsStorage(true)
            .pCmpTmp1(phaseSize)
            .pCmpTmp2(totalAddress)
            .pCmpTmp3(totalKeys)
            .pCmpTmp4(tupleSize)
            .pCmpTmp5(highPart(address))
            .pCmpTmp6(lowPart(address))
            .pCmpTmp7(totalStorageForThisAddress)
            .phaseEnd(phaseSize == 0);
        tracePostValues(trace, tracedValues);
      }
    }
  }
}
