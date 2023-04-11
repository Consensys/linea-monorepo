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
package net.consensys.linea.zktracer.module.alu.mod;

import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.bytes.MutableBytes32;

public class BaseBytes {
  protected final int OFFSET = 8;
  private final int LOW_HIGH_SIZE = 16;
  protected MutableBytes32 bytes32;

  static BaseBytes fromBytes32(Bytes32 arg){
    return new BaseBytes(arg);
  }

  protected BaseBytes(final Bytes32 arg) {
    bytes32 = arg.mutableCopy();
  }

  public Bytes16 getHigh() {
    return Bytes16.wrap(bytes32.slice(0, LOW_HIGH_SIZE));
  }

  public Bytes16 getLow() {
    return Bytes16.wrap(bytes32.slice(LOW_HIGH_SIZE));
  }

  public byte getByte(int index){
    return bytes32.get(index);
  }

  public Bytes32 getBytes32(){
    return bytes32;
  }
}
