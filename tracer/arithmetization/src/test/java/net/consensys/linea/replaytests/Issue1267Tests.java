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

import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Disabled
@Tag("nightly")
@Tag("replay")
public class Issue1267Tests {

  // splitting of 3506963-3507013
  @Test
  void issue_3506963_3506964() {
    replay(LINEA_MAINNET, "3506963-3506964.json.gz");
  }

  @Test
  void issue_3506965_3506969() {
    replay(LINEA_MAINNET, "3506965-3506969.json.gz");
  }

  @Test
  void issue_3506970_3506974() {
    replay(LINEA_MAINNET, "3506970-3506974.json.gz");
  }

  @Test
  void issue_3506975_3506979() {
    replay(LINEA_MAINNET, "3506975-3506979.json.gz");
  }

  @Test
  void issue_3506980_3506984() {
    replay(LINEA_MAINNET, "3506980-3506984.json.gz");
  }

  @Test
  void issue_3506985_3506989() {
    replay(LINEA_MAINNET, "3506985-3506989.json.gz");
  }

  @Test
  void issue_3506990_3506994() {
    replay(LINEA_MAINNET, "3506990-3506994.json.gz");
  }

  @Test
  void issue_3506995_3506999() {
    replay(LINEA_MAINNET, "3506995-3506999.json.gz");
  }

  @Test
  void issue_3507000_3507004() {
    replay(LINEA_MAINNET, "3507000-3507004.json.gz");
  }

  @Test
  void issue_3507005_3507009() {
    replay(LINEA_MAINNET, "3507005-3507009.json.gz");
  }

  @Test
  void issue_3507010_3507013() {
    replay(LINEA_MAINNET, "3507010-3507013.json.gz");
  }

  // splitting of 4065349-4065420
  @Test
  void issue_4065349_4065349() {
    replay(LINEA_MAINNET, "4065349-4065349.json.gz");
  }

  @Test
  void issue_4065350_4065354() {
    replay(LINEA_MAINNET, "4065350-4065354.json.gz");
  }

  @Test
  void issue_4065355_4065359() {
    replay(LINEA_MAINNET, "4065355-4065359.json.gz");
  }

  @Test
  void issue_4065360_4065364() {
    replay(LINEA_MAINNET, "4065360-4065364.json.gz");
  }

  // the only test that fails for me ... and only if I set resultChecking to true
  @Test
  void issue_4065365_4065369() {
    replay(LINEA_MAINNET, "4065365-4065369.json.gz", false);
  }

  @Test
  void issue_4065370_4065374() {
    replay(LINEA_MAINNET, "4065370-4065374.json.gz");
  }

  @Test
  void issue_4065375_4065379() {
    replay(LINEA_MAINNET, "4065375-4065379.json.gz");
  }

  @Test
  void issue_4065380_4065384() {
    replay(LINEA_MAINNET, "4065380-4065384.json.gz");
  }

  @Test
  void issue_4065385_4065389() {
    replay(LINEA_MAINNET, "4065385-4065389.json.gz");
  }

  @Test
  void issue_4065390_4065394() {
    replay(LINEA_MAINNET, "4065390-4065394.json.gz");
  }

  @Test
  void issue_4065395_4065399() {
    replay(LINEA_MAINNET, "4065395-4065399.json.gz");
  }

  @Test
  void issue_4065400_4065404() {
    replay(LINEA_MAINNET, "4065400-4065404.json.gz");
  }

  @Test
  void issue_4065405_4065409() {
    replay(LINEA_MAINNET, "4065405-4065409.json.gz");
  }

  @Test
  void issue_4065410_4065414() {
    replay(LINEA_MAINNET, "4065410-4065414.json.gz");
  }

  @Test
  void issue_4065415_4065419() {
    replay(LINEA_MAINNET, "4065415-4065419.json.gz");
  }

  @Test
  void issue_4065420_4065420() {
    replay(LINEA_MAINNET, "4065420-4065420.json.gz");
  }

  // splitting of 4736791-4736859
  @Test
  void issue_4736791_4736794() {
    replay(LINEA_MAINNET, "4736791-4736794.json.gz");
  }

  @Test
  void issue_4736795_4736799() {
    replay(LINEA_MAINNET, "4736795-4736799.json.gz");
  }

  @Test
  void issue_4736800_4736804() {
    replay(LINEA_MAINNET, "4736800-4736804.json.gz");
  }

  @Test
  void issue_4736805_4736809() {
    replay(LINEA_MAINNET, "4736805-4736809.json.gz");
  }

  @Test
  void issue_4736810_4736814() {
    replay(LINEA_MAINNET, "4736810-4736814.json.gz");
  }

  @Test
  void issue_4736815_4736819() {
    replay(LINEA_MAINNET, "4736815-4736819.json.gz");
  }

  @Test
  void issue_4736820_4736824() {
    replay(LINEA_MAINNET, "4736820-4736824.json.gz");
  }

  @Test
  void issue_4736825_4736829() {
    replay(LINEA_MAINNET, "4736825-4736829.json.gz");
  }

  @Test
  void issue_4736830_4736834() {
    replay(LINEA_MAINNET, "4736830-4736834.json.gz");
  }

  @Test
  void issue_4736835_4736839() {
    replay(LINEA_MAINNET, "4736835-4736839.json.gz");
  }

  @Test
  void issue_4736840_4736844() {
    replay(LINEA_MAINNET, "4736840-4736844.json.gz");
  }

  @Test
  void issue_4736845_4736849() {
    replay(LINEA_MAINNET, "4736845-4736849.json.gz");
  }

  @Test
  void issue_4736850_4736854() {
    replay(LINEA_MAINNET, "4736850-4736854.json.gz");
  }

  @Test
  void issue_4736855_4736859() {
    replay(LINEA_MAINNET, "4736855-4736859.json.gz");
  }
}
