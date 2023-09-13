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

package net.consensys.linea.zktracer.module.hub;

import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.Code;

public class Bytecode {
  public static Bytecode EMPTY = new Bytecode(Bytes.EMPTY);

  private final Bytes bytecode;
  private Hash hash;

  public Bytecode(Bytes bytes) {
    this.bytecode = bytes;
  }

  public Bytecode(Bytes bytes, Hash hash) {
    this.bytecode = bytes;
    this.hash = hash;
  }

  public Bytecode(Code code) {
    this.bytecode = code.getBytes();
    this.hash = code.getCodeHash();
  }

  public int getSize() {
    return this.bytecode.size();
  }

  public Bytes getBytes() {
    return this.bytecode;
  }

  public boolean isEmpty() {
    return this.bytecode.isEmpty();
  }

  public Hash getCodeHash() {
    if (this.hash == null) {
      this.hash = Hash.hash(this.bytecode);
    }
    return this.hash;
  }
}
