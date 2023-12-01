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

package net.consensys.linea.zktracer.module.stp;

import java.util.Optional;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;

public record StpChunk(
    OpCode opCode,
    Long gasActual,
    Long gasPrelim,
    Boolean oogx,
    Long gasMxp,
    Wei balance,
    Address to,
    Bytes32 value, // from stack
    // for Call only
    Optional<Boolean> toExists,
    Optional<Boolean> toWarm,
    Optional<Bytes32> gas) // from stack
{
  // Used by Create's instruction
  public StpChunk(
      OpCode opcode,
      Long gasActual,
      Long gasPrelim,
      Boolean oogx,
      Long gasMxp,
      Wei balance,
      Address to,
      Bytes32 value) {
    this(
        opcode,
        gasActual,
        gasPrelim,
        oogx,
        gasMxp,
        balance,
        to,
        value,
        Optional.empty(),
        Optional.empty(),
        Optional.empty());
  }

  // Used by Call's instruction
  public StpChunk(
      OpCode opcode,
      Long gasActual,
      Long gasPrelim,
      Boolean oogx,
      Long gasMxp,
      Wei balance,
      Address to,
      Bytes32 value,
      Boolean toExists,
      Boolean toWarm,
      Bytes32 gas) {
    this(
        opcode,
        gasActual,
        gasPrelim,
        oogx,
        gasMxp,
        balance,
        to,
        value,
        Optional.of(toExists),
        Optional.of(toWarm),
        Optional.of(gas));
  }
}
