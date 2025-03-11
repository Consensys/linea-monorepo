/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.core

import kotlin.random.Random
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever

class BeaconBlockHeaderTest {
  @Test
  fun `hash is not initialised on header construction`() {
    val headerHashFunction = Mockito.mock(HeaderHashFunction::class.java)

    val header = DataGenerators.randomBeaconBlockHeader(1u).copy(headerHashFunction = headerHashFunction)
    verify(headerHashFunction, Mockito.never()).invoke(header)
  }

  @Test
  fun `hash is calculated only once`() {
    val headerHashFunction = Mockito.mock(HeaderHashFunction::class.java)
    val header = DataGenerators.randomBeaconBlockHeader(1u).copy(headerHashFunction = headerHashFunction)
    whenever(headerHashFunction.invoke(header)).thenReturn(Random.nextBytes(32))

    verify(headerHashFunction, Mockito.never()).invoke(header)

    val hash1 = header.hash()
    val hash2 = header.hash()
    assertThat(hash1).isEqualTo(hash2)
    verify(headerHashFunction, times(1)).invoke(header)
  }
}
