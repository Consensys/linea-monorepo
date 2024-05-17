/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.OOB_INST_BLAKE_PARAMS;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.Blake2fMetadata;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import org.apache.tuweni.bytes.Bytes;

public record Blake2FPrecompile2(PrecompileInvocation p) implements OobCall {

  @Override
  public Bytes data(OobDataChannel i) {
    final Blake2fMetadata metadata = (Blake2fMetadata) p.metadata();
    return switch (i) {
      case DATA_1 -> Bytes.ofUnsignedLong(p.gasAtCall());
      case DATA_4 -> booleanToBytes(p.ramSuccess());
      case DATA_5 -> Bytes.ofUnsignedLong(p.ramSuccess() ? p.precompilePrice() - p.gasAtCall() : 0);
      case DATA_6 -> Bytes.ofUnsignedLong(metadata.r());
      case DATA_7 -> Bytes.ofUnsignedLong(metadata.f());

      default -> Bytes.EMPTY;
    };
  }

  @Override
  public int oobInstruction() {
    return OOB_INST_BLAKE_PARAMS;
  }
}
