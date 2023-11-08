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
package net.consensys.linea.zktracer.module.trm;

import static net.consensys.linea.zktracer.module.trm.Trm.isPrec;
import static org.assertj.core.api.Assertions.assertThat;

import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

public class TrmDataTest {

  @Test
  public void isPrecompile() {
    assertThat(isPrec(EWord.ZERO)).isFalse();
    assertThat(isPrec(EWord.ofHexString("0x1234"))).isFalse();
    // 0x06
    assertThat(isPrec(EWord.of(Address.ALTBN128_ADD))).isTrue();
    assertThat(isPrec(EWord.of(Address.MODEXP))).isTrue();
    // 0x09
    assertThat(isPrec(EWord.of(Address.BLAKE2B_F_COMPRESSION))).isTrue();
    // 0x01
    assertThat(isPrec(EWord.of(Address.ECREC))).isTrue();
    // 0x0A
    assertThat(isPrec(EWord.of(Address.BLS12_G1ADD))).isFalse(); // only true for 1-9
  }
}
