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
package linea.plugin.acc.test.compress;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import net.consensys.linea.compress.LibCompress;
import org.hyperledger.besu.tests.acceptance.dsl.AcceptanceTestBase;
import org.junit.jupiter.api.Test;

public class CompressTest extends AcceptanceTestBase {

  @Test
  public void testCompress() {
    final int compressedSize = LibCompress.CompressedSize(new byte[128], 128);

    // should not error
    assertThat(compressedSize).isGreaterThan(0);

    // should have compressed into 1 backref + header, must be less than 10
    assertThat(compressedSize).isLessThan(10);
  }
}
