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
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class BaseTheta  extends BaseBytes{

  private BaseTheta(final Bytes32 arg) {
    super(arg);
    bytes32 = arg.mutableCopy();
    for (int k = 0; k < 4; k++) {
      Bytes bytes = arg.slice(OFFSET * k, OFFSET);
      setBytes( OFFSET * (3-k) , bytes);
    }
  }

  static BaseTheta fromBytes32(Bytes32 arg){
    return new BaseTheta(arg);
  }
  public void setBytes(int index, Bytes bytes){
    bytes32.set(index, bytes);
  }

  public Bytes get(int index) {
    return bytes32.slice(OFFSET * index, OFFSET);
  }

  public Bytes slice (int i, int length){
    return bytes32.slice(i, length);
  }

  @Override
  public Bytes16 getHigh() {
    return Bytes16.wrap(Bytes.concatenate(get(3), get(2)));
  }
  @Override
  public Bytes16 getLow() {
    return Bytes16.wrap(Bytes.concatenate(get(1), get(0)));
  }
}
