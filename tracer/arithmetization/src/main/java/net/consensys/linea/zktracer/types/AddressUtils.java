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

import java.util.List;

import net.consensys.linea.zktracer.module.hub.transients.OperationAncillaries;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.Hash;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class AddressUtils {
  private static final Bytes CREATE2_PREFIX = Bytes.of(0xff);
  private static final List<Address> precompileAddress =
      List.of(
          Address.ECREC,
          Address.SHA256,
          Address.RIPEMD160,
          Address.ID,
          Address.MODEXP,
          Address.ALTBN128_ADD,
          Address.ALTBN128_MUL,
          Address.ALTBN128_PAIRING,
          Address.BLAKE2B_F_COMPRESSION);

  public static boolean isPrecompile(Address to) {
    return precompileAddress.contains(to);
  }

  /**
   * Compute the effective address of a transaction target, i.e. the specified target if explicitly
   * set, or the to-be-deployed address otherwise.
   *
   * @return the effective target address of tx
   */
  public static Address effectiveToAddress(Transaction tx) {
    return tx.getTo()
        .map(x -> (Address) x)
        .orElse(Address.contractAddress(tx.getSender(), tx.getNonce()));
  }

  public static Address getCreateAddress(final MessageFrame frame) {
    final Address currentAddress = frame.getRecipientAddress();
    return Address.contractAddress(
        currentAddress, frame.getWorldUpdater().get(currentAddress).getNonce());
  }

  public static Address getCreate2Address(final MessageFrame frame) {
    final Address sender = frame.getRecipientAddress();
    final Bytes32 salt = Bytes32.leftPad(frame.getStackItem(3));
    final Bytes initCode = OperationAncillaries.callData(frame);
    final Bytes32 hash =
        Hash.keccak256(Bytes.concatenate(CREATE2_PREFIX, sender, salt, Hash.keccak256(initCode)));
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
