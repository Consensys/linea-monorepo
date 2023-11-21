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

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.types.EWord;
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
    EWord eAddress = EWord.of(address);

    return trace
        .peekAtStorage(true)
        .pStorageAddressHi(eAddress.hiBigInt())
        .pStorageAddressLo(eAddress.loBigInt())
        .pStorageDeploymentNumber(BigInteger.valueOf(deploymentNumber))
        .pStorageStorageKeyHi(key.hiBigInt())
        .pStorageStorageKeyLo(key.loBigInt())
        .pStorageValOrigHi(valOrig.hiBigInt())
        .pStorageValOrigLo(valOrig.loBigInt())
        .pStorageValCurrHi(valCurr.hiBigInt())
        .pStorageValCurrLo(valCurr.loBigInt())
        .pStorageValNextHi(valNext.hiBigInt())
        .pStorageValNextLo(valNext.loBigInt())
        .pStorageWarm(oldWarmth)
        .pStorageWarmNew(newWarmth)
        .pStorageValOrigIsZero(valOrig.isZero())
        .pStorageValCurrIsZero(valCurr.isZero())
        .pStorageValNextIsZero(valNext.isZero())
        .pStorageValNextIsOrig(valNext == valOrig)
        .pStorageValNextIsCurr(valNext == valOrig);
  }
}
