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

package net.consensys.linea.zktracer.module.blake2f;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeComponents;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.crypto.Hash;

import java.util.Optional;

import static net.consensys.linea.zktracer.types.Conversions.bytesToBoolean;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class Blake2fOperation extends ModuleOperation {
  public static final short BLAKE2f_H_INPUT_CHUNK_SIZE = 8;
  public static final short BLAKE2f_M_CHUNK_SIZE = 8;
  static final short BLAKE2f_T_SIZE = 8;

  public final BlakeComponents blake2fComponents;

  public Blake2fOperation(BlakeComponents blakeComponents) {
      this.blake2fComponents = blakeComponents;
  }

  void trace(Trace.Blake2f trace) {
    Bytes result = Hash.blake2bf(blake2fComponents.callData());
    trace.r(blake2fComponents.r())
      .h0Input(blake2fComponents.getHashInput().slice(0, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h1Input(blake2fComponents.getHashInput().slice(8, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h2Input(blake2fComponents.getHashInput().slice(16, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h3Input(blake2fComponents.getHashInput().slice(24, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h4Input(blake2fComponents.getHashInput().slice(32, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h5Input(blake2fComponents.getHashInput().slice(40, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h6Input(blake2fComponents.getHashInput().slice(48, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h7Input(blake2fComponents.getHashInput().slice(56, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .m0(blake2fComponents.getHashInput().slice(64, BLAKE2f_M_CHUNK_SIZE))
      .m1(blake2fComponents.getHashInput().slice(72, BLAKE2f_M_CHUNK_SIZE))
      .m2(blake2fComponents.getHashInput().slice(80, BLAKE2f_M_CHUNK_SIZE))
      .m3(blake2fComponents.getHashInput().slice(88, BLAKE2f_M_CHUNK_SIZE))
      .m4(blake2fComponents.getHashInput().slice(96, BLAKE2f_M_CHUNK_SIZE))
      .m5(blake2fComponents.getHashInput().slice(104, BLAKE2f_M_CHUNK_SIZE))
      .m6(blake2fComponents.getHashInput().slice(112, BLAKE2f_M_CHUNK_SIZE))
      .m7(blake2fComponents.getHashInput().slice(120, BLAKE2f_M_CHUNK_SIZE))
      .m8(blake2fComponents.getHashInput().slice(128, BLAKE2f_M_CHUNK_SIZE))
      .m9(blake2fComponents.getHashInput().slice(136, BLAKE2f_M_CHUNK_SIZE))
      .m10(blake2fComponents.getHashInput().slice(144, BLAKE2f_M_CHUNK_SIZE))
      .m11(blake2fComponents.getHashInput().slice(152, BLAKE2f_M_CHUNK_SIZE))
      .m12(blake2fComponents.getHashInput().slice(160, BLAKE2f_M_CHUNK_SIZE))
      .m13(blake2fComponents.getHashInput().slice(168, BLAKE2f_M_CHUNK_SIZE))
      .m14(blake2fComponents.getHashInput().slice(176, BLAKE2f_M_CHUNK_SIZE))
      .m15(blake2fComponents.getHashInput().slice(184, BLAKE2f_M_CHUNK_SIZE))
      .t0(blake2fComponents.getHashInput().slice(192, BLAKE2f_T_SIZE))
      .t1(blake2fComponents.getHashInput().slice(200, BLAKE2f_T_SIZE))
      .f(bytesToBoolean(blake2fComponents.f()))
      .h0(result.slice(0, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h1(result.slice(8, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h2(result.slice(16, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h3(result.slice(24, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h4(result.slice(32, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h5(result.slice(40, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h6(result.slice(48, BLAKE2f_H_INPUT_CHUNK_SIZE))
      .h7(result.slice(56, BLAKE2f_H_INPUT_CHUNK_SIZE));


 }

  @Override
  protected int computeLineCount() {
    return 1;
  }
}
