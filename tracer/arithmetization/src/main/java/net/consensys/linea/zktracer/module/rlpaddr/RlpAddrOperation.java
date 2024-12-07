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

import static net.consensys.linea.zktracer.module.rlpaddr.Trace.MAX_CT_CREATE;
import static net.consensys.linea.zktracer.module.rlpaddr.Trace.MAX_CT_CREATE2;

import java.math.BigInteger;

import lombok.EqualsAndHashCode;
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
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public final class RlpAddrOperation extends ModuleOperation {
  @EqualsAndHashCode.Include private final Bytes32 rawHash;
  private final OpCode opCode;
  @EqualsAndHashCode.Include private final BigInteger nonce;
  private final Address address;
  private final Bytes32 salt;
  private final Bytes32 keccak;

  // CREATE operation
  public RlpAddrOperation(Bytes32 rawDepAddress, OpCode opCode, BigInteger nonce, Address address) {
    this(rawDepAddress, opCode, nonce, address, Bytes32.ZERO, Bytes32.ZERO);
  }

  // CREATE2 operation
  public RlpAddrOperation(
      Bytes32 rawHash,
      OpCode opCode,
      Address address,
      Bytes32 salt,
      Bytes32 kec,
      BigInteger nonce) {
    this(rawHash, opCode, nonce, address, salt, kec);
  }

  @Override
  protected int computeLineCount() {
    return this.opCode.equals(OpCode.CREATE) ? MAX_CT_CREATE + 1 : MAX_CT_CREATE2 + 1;
  }
}
