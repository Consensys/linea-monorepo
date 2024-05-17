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

package net.consensys.linea.zktracer.module.hub.fragment;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

/**
 * @param address target storage address
 * @param deploymentNumber storage account deployment number
 * @param key target storage slot
 * @param valOrig value @key at the beginning of *transaction*
 * @param valCurr value @key at the beginning of *opcode*
 */
public record StorageFragment(
    Address address,
    int deploymentNumber,
    EWord key,
    EWord valOrig,
    EWord valCurr,
    EWord valNext,
    boolean oldWarmth,
    boolean newWarmth)
    implements TraceFragment {
  @Override
  public Trace trace(Trace trace) {
    final EWord eAddress = EWord.of(address);

    return trace
        .peekAtStorage(true)
        .pStorageAddressHi(eAddress.hi())
        .pStorageAddressLo(eAddress.lo())
        .pStorageDeploymentNumber(Bytes.ofUnsignedInt(deploymentNumber))
        .pStorageStorageKeyHi(key.hi())
        .pStorageStorageKeyLo(key.lo())
        .pStorageValueOrigHi(valOrig.hi())
        .pStorageValueOrigLo(valOrig.lo())
        .pStorageValueCurrHi(valCurr.hi())
        .pStorageValueCurrLo(valCurr.lo())
        .pStorageValueNextHi(valNext.hi())
        .pStorageValueNextLo(valNext.lo())
        .pStorageWarmth(oldWarmth)
        .pStorageWarmthNew(newWarmth)
        .pStorageValueOrigIsZero(valOrig.isZero())
        .pStorageValueCurrIsZero(valCurr.isZero())
        .pStorageValueNextIsZero(valNext.isZero())
        .pStorageValueNextIsOrig(valNext == valOrig)
        .pStorageValueNextIsCurr(valNext == valOrig);
  }
}
