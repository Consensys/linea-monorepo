/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.rlpAuth;

import java.math.BigInteger;
import lombok.Getter;
import lombok.experimental.Accessors;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;

@Accessors(fluent = true)
@Getter
public class CodeDelegationTuple implements CodeDelegation {
  final BigInteger chainId;
  final Address address;
  final SECPSignature signature;
  final Address authorizer;
  final long nonce;
  final byte v;
  final BigInteger r;
  final BigInteger s;
  final long yParity;

  public CodeDelegationTuple(
      BigInteger chainId,
      Address address,
      SECPSignature signature,
      long nonce,
      byte v,
      BigInteger r,
      BigInteger s) {
    this.chainId = chainId;
    this.address = address;
    this.signature = signature;
    this.authorizer = Address.ALTBN128_ADD; // placeholder
    // authority = ecrecover(msg, y_parity, r, s)
    // msg = keccak(MAGIC || rlp([chain_id, address, nonce])).
    this.nonce = nonce;
    this.v = v;
    this.r = r;
    this.s = s;
    this.yParity = v == 28 ? 1L : 0L;
  }
}
