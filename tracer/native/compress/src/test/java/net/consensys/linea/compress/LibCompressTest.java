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
 *
 */
package net.consensys.linea.compress;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

public class LibCompressTest {

    @Test
    public void testCompressZeroes() {
        byte[] zeroes = new byte[128];
        int size = LibCompress.CompressedSize(zeroes, 128);

        // should not error
        assertThat(size).isGreaterThan(0);

        // should have compressed into 1 backref + header, must be less than 10
        assertThat(size).isLessThan(10);
    }

    @Test
    public void testCompressTooLargeInput() {
        byte[] zeroes = new byte[512*1024];
        int size = LibCompress.CompressedSize(zeroes, 512*1024);

        // should error --> too large payload.
        assertThat(size).isLessThan(0);
    }

}