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

@Tag("nightly")
@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class Issue1280Tests {

  // 3901959-3902032
  @Test
  void issue1280_range_3901959_3901959() {
    replay(OLD_MAINNET_TESTCONFIG, "3901959.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901960_3901964() {
    replay(OLD_MAINNET_TESTCONFIG, "3901960-3901964.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901965_3901969() {
    replay(OLD_MAINNET_TESTCONFIG, "3901965-3901969.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901970_3901974() {
    replay(OLD_MAINNET_TESTCONFIG, "3901970-3901974.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901975_3901979() {
    replay(OLD_MAINNET_TESTCONFIG, "3901975-3901979.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901980_3901984() {
    replay(OLD_MAINNET_TESTCONFIG, "3901980-3901984.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901985_3901989() {
    replay(OLD_MAINNET_TESTCONFIG, "3901985-3901989.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901990_3901994() {
    replay(OLD_MAINNET_TESTCONFIG, "3901990-3901994.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3901995_3901999() {
    replay(OLD_MAINNET_TESTCONFIG, "3901995-3901999.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902000_3902004() {
    replay(OLD_MAINNET_TESTCONFIG, "3902000-3902004.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902005_3902009() {
    replay(OLD_MAINNET_TESTCONFIG, "3902005-3902009.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902010_3902014() {
    replay(OLD_MAINNET_TESTCONFIG, "3902010-3902014.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902015_3902019() {
    replay(OLD_MAINNET_TESTCONFIG, "3902015-3902019.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902020_3902024() {
    replay(OLD_MAINNET_TESTCONFIG, "3902020-3902024.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902025_3902029() {
    replay(OLD_MAINNET_TESTCONFIG, "3902025-3902029.mainnet.json.gz");
  }

  @Test
  void issue1280_range_3902030_3902032() {
    replay(OLD_MAINNET_TESTCONFIG, "3902030-3902032.mainnet.json.gz");
  }

  // 4065349-4065420
  @Test
  void issue1280_range_4065349_4065349() {
    replay(OLD_MAINNET_TESTCONFIG, "4065349.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065350_4065354() {
    replay(OLD_MAINNET_TESTCONFIG, "4065350-4065354.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065355_4065359() {
    replay(OLD_MAINNET_TESTCONFIG, "4065355-4065359.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065360_4065364() {
    replay(OLD_MAINNET_TESTCONFIG, "4065360-4065364.mainnet.json.gz");
  }

  /** this range fails if you use the default resultChecking == true */
  @Test
  void issue1280_range_4065365_4065369() {
    replay(OLD_MAINNET_TESTCONFIG, "4065365-4065369.mainnet.json.gz", false);
  }

  @Test
  void issue1280_range_4065370_4065374() {
    replay(OLD_MAINNET_TESTCONFIG, "4065370-4065374.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065375_4065379() {
    replay(OLD_MAINNET_TESTCONFIG, "4065375-4065379.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065380_4065384() {
    replay(OLD_MAINNET_TESTCONFIG, "4065380-4065384.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065385_4065389() {
    replay(OLD_MAINNET_TESTCONFIG, "4065385-4065389.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065390_4065394() {
    replay(OLD_MAINNET_TESTCONFIG, "4065390-4065394.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065395_4065399() {
    replay(OLD_MAINNET_TESTCONFIG, "4065395-4065399.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065400_4065404() {
    replay(OLD_MAINNET_TESTCONFIG, "4065400-4065404.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065405_4065409() {
    replay(OLD_MAINNET_TESTCONFIG, "4065405-4065409.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065410_4065414() {
    replay(OLD_MAINNET_TESTCONFIG, "4065410-4065414.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065415_4065419() {
    replay(OLD_MAINNET_TESTCONFIG, "4065415-4065419.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4065420_4065420() {
    replay(OLD_MAINNET_TESTCONFIG, "4065420.mainnet.json.gz");
  }

  // 4736791-4736859
  @Test
  void issue1280_range_4736791_4736794() {
    replay(OLD_MAINNET_TESTCONFIG, "4736791-4736794.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736795_4736799() {
    replay(OLD_MAINNET_TESTCONFIG, "4736795-4736799.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736800_4736804() {
    replay(OLD_MAINNET_TESTCONFIG, "4736800-4736804.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736805_4736809() {
    replay(OLD_MAINNET_TESTCONFIG, "4736805-4736809.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736810_4736814() {
    replay(OLD_MAINNET_TESTCONFIG, "4736810-4736814.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736815_4736819() {
    replay(OLD_MAINNET_TESTCONFIG, "4736815-4736819.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736820_4736824() {
    replay(OLD_MAINNET_TESTCONFIG, "4736820-4736824.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736825_4736829() {
    replay(OLD_MAINNET_TESTCONFIG, "4736825-4736829.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736830_4736834() {
    replay(OLD_MAINNET_TESTCONFIG, "4736830-4736834.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736835_4736839() {
    replay(OLD_MAINNET_TESTCONFIG, "4736835-4736839.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736840_4736844() {
    replay(OLD_MAINNET_TESTCONFIG, "4736840-4736844.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736845_4736849() {
    replay(OLD_MAINNET_TESTCONFIG, "4736845-4736849.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736850_4736854() {
    replay(OLD_MAINNET_TESTCONFIG, "4736850-4736854.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4736855_4736859() {
    replay(OLD_MAINNET_TESTCONFIG, "4736855-4736859.mainnet.json.gz");
  }

  // 4981619-4981658
  @Test
  void issue1280_range_4981619_4981619() {
    replay(OLD_MAINNET_TESTCONFIG, "4981619.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981620_4981624() {
    replay(OLD_MAINNET_TESTCONFIG, "4981620-4981624.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981625_4981629() {
    replay(OLD_MAINNET_TESTCONFIG, "4981625-4981629.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981630_4981634() {
    replay(OLD_MAINNET_TESTCONFIG, "4981630-4981634.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981635_4981639() {
    replay(OLD_MAINNET_TESTCONFIG, "4981635-4981639.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981640_4981644() {
    replay(OLD_MAINNET_TESTCONFIG, "4981640-4981644.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981645_4981649() {
    replay(OLD_MAINNET_TESTCONFIG, "4981645-4981649.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981650_4981654() {
    replay(OLD_MAINNET_TESTCONFIG, "4981650-4981654.mainnet.json.gz");
  }

  @Test
  void issue1280_range_4981655_4981658() {
    replay(OLD_MAINNET_TESTCONFIG, "4981655-4981658.mainnet.json.gz");
  }
}
