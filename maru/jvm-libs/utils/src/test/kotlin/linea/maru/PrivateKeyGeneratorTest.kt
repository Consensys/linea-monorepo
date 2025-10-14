/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.maru

import linea.maru.PrivateKeyGenerator.generatePrivateKey
import linea.maru.PrivateKeyGenerator.getKeyData
import linea.maru.PrivateKeyGenerator.getKeyDataByPrefixedKey
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.RepeatedTest

class PrivateKeyGeneratorTest {
  @RepeatedTest(100)
  fun `should generator private key`() {
    val keyData = generatePrivateKey()
    val recoveredDataFromPrivKey = getKeyData(privateKey = keyData.privateKey)
    val recoveredDataFromPrivKeyPrefixed = getKeyDataByPrefixedKey(prefixedPrivateKey = keyData.prefixedPrivateKey)
    assertThat(keyData).isEqualTo(recoveredDataFromPrivKey)
    assertThat(keyData).isEqualTo(recoveredDataFromPrivKeyPrefixed)
  }
}
