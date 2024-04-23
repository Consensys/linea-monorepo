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

package net.consensys.linea.zktracer.module.rlpaddr;

import java.math.BigInteger;
import java.util.Optional;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;

@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public final class RlpAddrChunk extends ModuleOperation {
  private final Address depAddress;
  private final OpCode opCode;
  private final Optional<BigInteger> nonce;
  private final Address address;
  private final Optional<Bytes32> salt;
  private final Optional<Bytes32> keccak;

  public RlpAddrChunk(Address depAddress, OpCode opCode, BigInteger nonce, Address address) {
    this(depAddress, opCode, Optional.of(nonce), address, Optional.empty(), Optional.empty());
  }

  public RlpAddrChunk(
      Address depAddress, OpCode opCode, Address address, Bytes32 salt, Bytes32 kec) {
    this(depAddress, opCode, Optional.empty(), address, Optional.of(salt), Optional.of(kec));
  }

  @Override
  protected int computeLineCount() {
    return this.opCode.equals(OpCode.CREATE) ? 8 : 6;
  }
}
