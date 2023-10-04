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

package net.consensys.linea.zktracer.module.rlpAddr;

import java.util.Optional;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;

public record RlpAddrChunk(
    OpCode opCode,
    Optional<Long> nonce,
    Address address,
    Optional<Bytes32> salt,
    Optional<Bytes32> keccak) {
  public RlpAddrChunk(OpCode opCode, long nonce, Address address) {
    this(opCode, Optional.of(nonce), address, Optional.empty(), Optional.empty());
  }

  public RlpAddrChunk(OpCode opCode, Address address, Bytes32 salt, Bytes32 kec) {
    this(opCode, Optional.empty(), address, Optional.of(salt), Optional.of(kec));
  }
}
