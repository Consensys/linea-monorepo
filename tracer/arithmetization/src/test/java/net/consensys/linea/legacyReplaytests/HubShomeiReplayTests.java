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

package net.consensys.linea.legacyReplaytests;

import static net.consensys.linea.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.MAINNET_LONDON_TESTCONFIG;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@Tag("replay")
@Tag("nightly")
@Disabled
@ExtendWith(UnitTestWatcher.class)
public class HubShomeiReplayTests extends TracerTestBase {

  /**
   * Address: 0x8d95f56b0bac46e8ac1d3a3f12fb1e5bc39b4c0c Storage key hi:
   * 0x4955ac1f8710286d713fc7cfabe0953 Storage key hi: 0xf7799e9ee340180bcdfd35c3d33c99a3 is in
   * Shomei, not in Hub. This is due to an SSTOREX (remaining gas < 2300), that happens in block 11.
   * In block 35 (tx number 53) we have a succesfull non reverted SSTORE at the same acc / key.
   */
  @Test
  void alert2025_06_12_first(TestInfo testInfo) {
    replay(MAINNET_LONDON_TESTCONFIG, "legacy/19913402-19913483.mainnet.json.gz", testInfo);
  }

  @Test
  void alert2025_06_12_second(TestInfo testInfo) {
    replay(MAINNET_LONDON_TESTCONFIG, "legacy/19914560-19914640.mainnet.json.gz", testInfo);
  }
}
