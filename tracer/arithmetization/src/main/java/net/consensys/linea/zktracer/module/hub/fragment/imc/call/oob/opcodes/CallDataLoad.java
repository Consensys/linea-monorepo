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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.opcodes;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.OOB_INST_CDL;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

public record CallDataLoad(EWord readOffset, EWord callDataSize) implements OobCall {
  public static CallDataLoad build(Hub hub, MessageFrame frame) {
    return new CallDataLoad(
        EWord.of(frame.getStackItem(0)), EWord.of(hub.currentFrame().callDataInfo().data().size()));
  }

  @Override
  public Bytes data(OobDataChannel i) {
    return switch (i) {
      case DATA_1 -> this.readOffset.hi();
      case DATA_2 -> this.readOffset.lo();
      case DATA_5 -> this.callDataSize;
      case DATA_7 -> booleanToBytes(this.readOffset.greaterOrEqualThan(this.callDataSize));
      default -> Bytes.EMPTY;
    };
  }

  @Override
  public int oobInstruction() {
    return OOB_INST_CDL;
  }
}
