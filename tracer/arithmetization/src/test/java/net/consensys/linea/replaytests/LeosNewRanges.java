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
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@Disabled
@ExtendWith(UnitTestWatcher.class)
@Tag("nightly")
public class LeosNewRanges {
  @Disabled
  @Test
  void leos_new_ranges_2258472_2258607() {
    replay(OLD_MAINNET_TESTCONFIG, "2258472-2258607.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_2291967_2292180() {
    replay(OLD_MAINNET_TESTCONFIG, "2291967-2292180.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_2321460_2321556() {
    replay(OLD_MAINNET_TESTCONFIG, "2321460-2321556.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_2359782_2359913() {
    replay(OLD_MAINNET_TESTCONFIG, "2359782-2359913.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_2362189_2362291() {
    replay(OLD_MAINNET_TESTCONFIG, "2362189-2362291.mainnet.json.gz");
  }

  @Disabled("Out-Of-Memory")
  @Test
  void leos_new_ranges_5002125_5002158() {
    replay(OLD_MAINNET_TESTCONFIG, "5002125-5002158.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5004016_5004055() {
    replay(OLD_MAINNET_TESTCONFIG, "5004016-5004055.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5004767_5004806() {
    replay(OLD_MAINNET_TESTCONFIG, "5004767-5004806.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5006057_5006092() {
    replay(OLD_MAINNET_TESTCONFIG, "5006057-5006092.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5006988_5007039() {
    replay(OLD_MAINNET_TESTCONFIG, "5006988-5007039.mainnet.json.gz");
  }

  @Disabled("Out-Of-Memory")
  @Test
  void leos_new_ranges_5012236_5012275() {
    replay(OLD_MAINNET_TESTCONFIG, "5012236-5012275.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5025817_5025859() {
    replay(OLD_MAINNET_TESTCONFIG, "5025817-5025859.mainnet.json.gz");
  }

  @Disabled("Out-Of-Memory")
  @Test
  void leos_new_ranges_5037583_5037608() {
    replay(OLD_MAINNET_TESTCONFIG, "5037583-5037608.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5042942_5042990() {
    replay(OLD_MAINNET_TESTCONFIG, "5042942-5042990.mainnet.json.gz");
  }

  @Disabled("Out-Of-Memory")
  @Test
  void leos_new_ranges_5043442_5043497() {
    replay(OLD_MAINNET_TESTCONFIG, "5043442-5043497.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5043997_5044049() {
    replay(OLD_MAINNET_TESTCONFIG, "5043997-5044049.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5044557_5044619() {
    replay(OLD_MAINNET_TESTCONFIG, "5044557-5044619.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5045161_5045232() {
    replay(OLD_MAINNET_TESTCONFIG, "5045161-5045232.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5046373_5046435() {
    replay(OLD_MAINNET_TESTCONFIG, "5046373-5046435.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5046997_5047058() {
    replay(OLD_MAINNET_TESTCONFIG, "5046997-5047058.mainnet.json.gz");
  }

  @Disabled
  @Test
  void leos_new_ranges_5050036_5050130() {
    replay(OLD_MAINNET_TESTCONFIG, "5050036-5050130.mainnet.json.gz");
  }

  @Disabled("Out-Of-Memory")
  @Test
  void leos_new_ranges_5057558_5057616() {
    replay(OLD_MAINNET_TESTCONFIG, "5057558-5057616.mainnet.json.gz");
  }
}
