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

package net.consensys.linea.zktracer;

import java.security.MessageDigest;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.bouncycastle.jcajce.provider.digest.Keccak;
import org.bouncycastle.jcajce.provider.digest.RIPEMD160;

/** Contains different types of hash function implementations. */
public class Hash {

  /**
   * Generates a {@link Keccak} digest.
   *
   * @param input bytes input
   * @return a {@link Bytes32} digest
   */
  public static Bytes32 keccak256(final Bytes input) {
    final MessageDigest digest = new Keccak.Digest256();
    input.update(digest);

    return Bytes32.wrap(digest.digest());
  }

  /**
   * Generates a {@link RIPEMD160} digest.
   *
   * @param input bytes input
   * @return a {@link Bytes32} digest
   */
  public static Bytes32 ripemd160(final Bytes input) {
    final MessageDigest digest = new RIPEMD160.Digest();
    input.update(digest);

    return Bytes32.wrap(digest.digest());
  }
}
