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

package net.consensys.linea.zktracer.types;

import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.Hash;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class AddressUtils {
  private static final List<Address> precompileAddress =
      List.of(
          org.hyperledger.besu.datatypes.Address.ECREC,
          org.hyperledger.besu.datatypes.Address.SHA256,
          org.hyperledger.besu.datatypes.Address.RIPEMD160,
          org.hyperledger.besu.datatypes.Address.ID,
          org.hyperledger.besu.datatypes.Address.MODEXP,
          org.hyperledger.besu.datatypes.Address.ALTBN128_ADD,
          org.hyperledger.besu.datatypes.Address.ALTBN128_MUL,
          org.hyperledger.besu.datatypes.Address.ALTBN128_PAIRING,
          org.hyperledger.besu.datatypes.Address.BLAKE2B_F_COMPRESSION);

  public static boolean isPrecompile(org.hyperledger.besu.datatypes.Address to) {
    return precompileAddress.contains(to);
  }

  public static Address getCreateAddress(final MessageFrame frame) {
    if (!OpCode.of(frame.getCurrentOperation().getOpcode()).equals(OpCode.CREATE)) {
      throw new IllegalArgumentException("Must be called only for CREATE opcode");
    }
    final Address currentAddress = frame.getRecipientAddress();
    return Address.contractAddress(
        currentAddress, frame.getWorldUpdater().get(currentAddress).getNonce());
  }

  public static Address getCreate2Address(final MessageFrame frame) {
    if (!OpCode.of(frame.getCurrentOperation().getOpcode()).equals(OpCode.CREATE2)) {
      throw new IllegalArgumentException("Must be called only for CREATE2 opcode");
    }
    final Address sender = frame.getRecipientAddress();
    final Bytes32 salt = Bytes32.leftPad(frame.getStackItem(3));
    final long offset = clampedToLong(frame.getStackItem(1));
    final long length = clampedToLong(frame.getStackItem(2));
    final Bytes initCode = frame.shadowReadMemory(offset, length);
    Bytes PREFIX = Bytes.fromHexString("0xff");
    final Bytes32 hash =
        Hash.keccak256(Bytes.concatenate(PREFIX, sender, salt, Hash.keccak256(initCode)));
    return Address.extract(hash);
  }

  public static Address getDeploymentAddress(final MessageFrame frame) {
    OpCode opcode = OpCode.of(frame.getCurrentOperation().getOpcode());
    if (!opcode.equals(OpCode.CREATE2) && !opcode.equals(OpCode.CREATE)) {
      throw new IllegalArgumentException("Must be called only for CREATE/CREATE2 opcode");
    }
    return opcode.equals(OpCode.CREATE) ? getCreateAddress(frame) : getCreate2Address(frame);
  }
}
