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
package net.consensys.linea.replaytests;

import static net.consensys.linea.replaytests.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.OLD_MAINNET_TESTCONFIG;

import net.consensys.linea.UnitTestWatcher;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class Issue1090Tests {

  /**
   * This is an interesting block: at transaction 13 (which is a CONTRACT_CREATION transaction) the
   * address being deployed at the ROOT does a DELEGATECALL to 0xa092..., which then CALLs its
   * CALLER address twice. The CALLER, being under deployment, is seen as having empty bytecode by
   * the outside world, in particular by 0xa092... .
   */
  @Test
  void issue_1090_block_1507291() {
    replay(OLD_MAINNET_TESTCONFIG, "1507291.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1809818() {
    replay(OLD_MAINNET_TESTCONFIG, "1809818.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1812784() {
    replay(OLD_MAINNET_TESTCONFIG, "1812784.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1814860() {
    replay(OLD_MAINNET_TESTCONFIG, "1814860.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1714851() {
    replay(OLD_MAINNET_TESTCONFIG, "1714851.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1505729() {
    replay(OLD_MAINNET_TESTCONFIG, "1505729.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1511808() {
    replay(OLD_MAINNET_TESTCONFIG, "1511808.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1400040() {
    replay(OLD_MAINNET_TESTCONFIG, "1400040.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1409462() {
    replay(OLD_MAINNET_TESTCONFIG, "1409462.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1410650() {
    replay(OLD_MAINNET_TESTCONFIG, "1410650.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1204298() {
    replay(OLD_MAINNET_TESTCONFIG, "1204298.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1213822() {
    replay(OLD_MAINNET_TESTCONFIG, "1213822.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1214117() {
    replay(OLD_MAINNET_TESTCONFIG, "1214117.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1214259() {
    replay(OLD_MAINNET_TESTCONFIG, "1214259.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1214280() {
    replay(OLD_MAINNET_TESTCONFIG, "1214280.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1214528() {
    replay(OLD_MAINNET_TESTCONFIG, "1214528.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1100579() {
    replay(OLD_MAINNET_TESTCONFIG, "1100579.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1104982() {
    replay(OLD_MAINNET_TESTCONFIG, "1104982.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1105022() {
    replay(OLD_MAINNET_TESTCONFIG, "1105022.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1105029() {
    replay(OLD_MAINNET_TESTCONFIG, "1105029.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1105038() {
    replay(OLD_MAINNET_TESTCONFIG, "1105038.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1106506() {
    replay(OLD_MAINNET_TESTCONFIG, "1106506.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1106648() {
    replay(OLD_MAINNET_TESTCONFIG, "1106648.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1107867() {
    replay(OLD_MAINNET_TESTCONFIG, "1107867.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1002387() {
    replay(OLD_MAINNET_TESTCONFIG, "1002387.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1003970() {
    replay(OLD_MAINNET_TESTCONFIG, "1003970.mainnet.json.gz");
  }

  @Tag("nightly")
  @Test
  void issue_1090_block_1010069() {
    replay(OLD_MAINNET_TESTCONFIG, "1010069.mainnet.json.gz");
  }
}
