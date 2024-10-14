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
import static net.consensys.linea.testing.ReplayExecutionEnvironment.LINEA_MAINNET;

import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("nightly")
@Tag("replay")
public class Issue1280Tests {

  // 3901959-3902032
  @Test
  void issue1280_range_3901959_3901959() {
    replay(LINEA_MAINNET, "3901959-3901959.json.gz");
  }

  @Test
  void issue1280_range_3901960_3901964() {
    replay(LINEA_MAINNET, "3901960-3901964.json.gz");
  }

  @Test
  void issue1280_range_3901965_3901969() {
    replay(LINEA_MAINNET, "3901965-3901969.json.gz");
  }

  @Test
  void issue1280_range_3901970_3901974() {
    replay(LINEA_MAINNET, "3901970-3901974.json.gz");
  }

  @Test
  void issue1280_range_3901975_3901979() {
    replay(LINEA_MAINNET, "3901975-3901979.json.gz");
  }

  @Test
  void issue1280_range_3901980_3901984() {
    replay(LINEA_MAINNET, "3901980-3901984.json.gz");
  }

  @Test
  void issue1280_range_3901985_3901989() {
    replay(LINEA_MAINNET, "3901985-3901989.json.gz");
  }

  @Test
  void issue1280_range_3901990_3901994() {
    replay(LINEA_MAINNET, "3901990-3901994.json.gz");
  }

  @Test
  void issue1280_range_3901995_3901999() {
    replay(LINEA_MAINNET, "3901995-3901999.json.gz");
  }

  @Test
  void issue1280_range_3902000_3902004() {
    replay(LINEA_MAINNET, "3902000-3902004.json.gz");
  }

  @Test
  void issue1280_range_3902005_3902009() {
    replay(LINEA_MAINNET, "3902005-3902009.json.gz");
  }

  @Test
  void issue1280_range_3902010_3902014() {
    replay(LINEA_MAINNET, "3902010-3902014.json.gz");
  }

  @Test
  void issue1280_range_3902015_3902019() {
    replay(LINEA_MAINNET, "3902015-3902019.json.gz");
  }

  @Test
  void issue1280_range_3902020_3902024() {
    replay(LINEA_MAINNET, "3902020-3902024.json.gz");
  }

  @Test
  void issue1280_range_3902025_3902029() {
    replay(LINEA_MAINNET, "3902025-3902029.json.gz");
  }

  @Test
  void issue1280_range_3902030_3902032() {
    replay(LINEA_MAINNET, "3902030-3902032.json.gz");
  }

  // 4065349-4065420
  @Test
  void issue1280_range_4065349_4065349() {
    replay(LINEA_MAINNET, "4065349-4065349.json.gz");
  }

  @Test
  void issue1280_range_4065350_4065354() {
    replay(LINEA_MAINNET, "4065350-4065354.json.gz");
  }

  @Test
  void issue1280_range_4065355_4065359() {
    replay(LINEA_MAINNET, "4065355-4065359.json.gz");
  }

  @Test
  void issue1280_range_4065360_4065364() {
    replay(LINEA_MAINNET, "4065360-4065364.json.gz");
  }

  /** this range fails if you use the default resultChecking == true */
  @Test
  void issue1280_range_4065365_4065369() {
    replay(LINEA_MAINNET, "4065365-4065369.json.gz", false);
  }

  @Test
  void issue1280_range_4065370_4065374() {
    replay(LINEA_MAINNET, "4065370-4065374.json.gz");
  }

  @Test
  void issue1280_range_4065375_4065379() {
    replay(LINEA_MAINNET, "4065375-4065379.json.gz");
  }

  @Test
  void issue1280_range_4065380_4065384() {
    replay(LINEA_MAINNET, "4065380-4065384.json.gz");
  }

  @Test
  void issue1280_range_4065385_4065389() {
    replay(LINEA_MAINNET, "4065385-4065389.json.gz");
  }

  @Test
  void issue1280_range_4065390_4065394() {
    replay(LINEA_MAINNET, "4065390-4065394.json.gz");
  }

  @Test
  void issue1280_range_4065395_4065399() {
    replay(LINEA_MAINNET, "4065395-4065399.json.gz");
  }

  @Test
  void issue1280_range_4065400_4065404() {
    replay(LINEA_MAINNET, "4065400-4065404.json.gz");
  }

  @Test
  void issue1280_range_4065405_4065409() {
    replay(LINEA_MAINNET, "4065405-4065409.json.gz");
  }

  @Test
  void issue1280_range_4065410_4065414() {
    replay(LINEA_MAINNET, "4065410-4065414.json.gz");
  }

  @Test
  void issue1280_range_4065415_4065419() {
    replay(LINEA_MAINNET, "4065415-4065419.json.gz");
  }

  @Test
  void issue1280_range_4065420_4065420() {
    replay(LINEA_MAINNET, "4065420-4065420.json.gz");
  }

  // 4736791-4736859
  @Test
  void issue1280_range_4736791_4736794() {
    replay(LINEA_MAINNET, "4736791-4736794.json.gz");
  }

  @Test
  void issue1280_range_4736795_4736799() {
    replay(LINEA_MAINNET, "4736795-4736799.json.gz");
  }

  @Test
  void issue1280_range_4736800_4736804() {
    replay(LINEA_MAINNET, "4736800-4736804.json.gz");
  }

  @Test
  void issue1280_range_4736805_4736809() {
    replay(LINEA_MAINNET, "4736805-4736809.json.gz");
  }

  @Test
  void issue1280_range_4736810_4736814() {
    replay(LINEA_MAINNET, "4736810-4736814.json.gz");
  }

  @Test
  void issue1280_range_4736815_4736819() {
    replay(LINEA_MAINNET, "4736815-4736819.json.gz");
  }

  @Test
  void issue1280_range_4736820_4736824() {
    replay(LINEA_MAINNET, "4736820-4736824.json.gz");
  }

  @Test
  void issue1280_range_4736825_4736829() {
    replay(LINEA_MAINNET, "4736825-4736829.json.gz");
  }

  @Test
  void issue1280_range_4736830_4736834() {
    replay(LINEA_MAINNET, "4736830-4736834.json.gz");
  }

  @Test
  void issue1280_range_4736835_4736839() {
    replay(LINEA_MAINNET, "4736835-4736839.json.gz");
  }

  @Test
  void issue1280_range_4736840_4736844() {
    replay(LINEA_MAINNET, "4736840-4736844.json.gz");
  }

  @Test
  void issue1280_range_4736845_4736849() {
    replay(LINEA_MAINNET, "4736845-4736849.json.gz");
  }

  @Test
  void issue1280_range_4736850_4736854() {
    replay(LINEA_MAINNET, "4736850-4736854.json.gz");
  }

  @Test
  void issue1280_range_4736855_4736859() {
    replay(LINEA_MAINNET, "4736855-4736859.json.gz");
  }

  // 4981619-4981658
  @Test
  void issue1280_range_4981619_4981619() {
    replay(LINEA_MAINNET, "4981619-4981619.json.gz");
  }

  @Test
  void issue1280_range_4981620_4981624() {
    replay(LINEA_MAINNET, "4981620-4981624.json.gz");
  }

  @Test
  void issue1280_range_4981625_4981629() {
    replay(LINEA_MAINNET, "4981625-4981629.json.gz");
  }

  @Test
  void issue1280_range_4981630_4981634() {
    replay(LINEA_MAINNET, "4981630-4981634.json.gz");
  }

  @Test
  void issue1280_range_4981635_4981639() {
    replay(LINEA_MAINNET, "4981635-4981639.json.gz");
  }

  @Test
  void issue1280_range_4981640_4981644() {
    replay(LINEA_MAINNET, "4981640-4981644.json.gz");
  }

  @Test
  void issue1280_range_4981645_4981649() {
    replay(LINEA_MAINNET, "4981645-4981649.json.gz");
  }

  @Test
  void issue1280_range_4981650_4981654() {
    replay(LINEA_MAINNET, "4981650-4981654.json.gz");
  }

  @Test
  void issue1280_range_4981655_4981658() {
    replay(LINEA_MAINNET, "4981655-4981658.json.gz");
  }
}
