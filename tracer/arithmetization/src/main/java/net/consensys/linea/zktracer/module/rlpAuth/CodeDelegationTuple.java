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

import static org.hyperledger.besu.crypto.Hash.keccak256;
import static org.hyperledger.besu.ethereum.core.CodeDelegation.MAGIC;

import java.math.BigInteger;
import java.util.Optional;
import lombok.Getter;
import lombok.experimental.Accessors;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.SECPPublicKey;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.ethereum.rlp.*;

@Accessors(fluent = true)
@Getter
public class CodeDelegationTuple implements CodeDelegation {
  final BigInteger chainId;
  final Address address;
  final SECPSignature signature;
  final Optional<Address> authorizer; // predicted output from ecRecover
  final long nonce;
  final byte v;
  final BigInteger r;
  final BigInteger s;
  final long yParity;

  final Bytes32 msg; // predicted output from keccak256
  boolean authorityEcRecoverSuccess;

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
    this.nonce = nonce;
    this.v = v;
    this.r = r;
    this.s = s;
    this.yParity = v == 28 ? 1L : 0L;

    // authority = ecrecover(msg, y_parity, r, s)
    // msg = keccak(MAGIC || rlp([chain_id, address, nonce]))

    final Bytes keccakInput =
        Bytes.concatenate(MAGIC, rlpOfListWithChainIdAddressNonce(chainId, address, nonce));

    this.msg = keccak256(keccakInput);
    this.authorizer = ecRecover(msg, yParity, r, s);
  }

  Bytes rlpOfListWithChainIdAddressNonce(BigInteger chainId, Address address, long nonce) {
    final BytesValueRLPOutput listRlp = new BytesValueRLPOutput();
    listRlp.startList();
    listRlp.writeBigIntegerScalar(chainId);
    listRlp.writeBytes(address);
    listRlp.writeLongScalar(nonce);
    listRlp.endList();
    return listRlp.encoded();
  }

  Optional<Address> ecRecover(
      final Bytes32 h, final long yParity, final BigInteger r, final BigInteger s) {
    final SignatureAlgorithm sigAlg = SignatureAlgorithmFactory.getInstance();
    final SECPSignature sig = new SECPSignature(r, s, (byte) yParity);
    final Optional<SECPPublicKey> pubKey = sigAlg.recoverPublicKeyFromSignature(h, sig);
    if (pubKey.isEmpty()) {
      this.authorityEcRecoverSuccess = false;
      return Optional.empty();
    }
    this.authorityEcRecoverSuccess = true;
    // The address is represented by the last 20 bytes of keccak256(uncompressedPubKey[1:])
    final Bytes uncompressedPubKey = pubKey.get().getEncodedBytes(); // 65 bytes, 0x04 || X || Y
    final Bytes32 hashed = keccak256(uncompressedPubKey.slice(1)); // Drop the 0x04 prefix
    return Optional.of(Address.wrap(hashed.slice(12, 20))); // Get the last 20 bytes
  }
}
