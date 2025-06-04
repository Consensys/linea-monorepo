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
package net.consensys.linea.reporting;

import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.Fork;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.TestInfo;

public class TracerTestBase {
  public static TestInfoWithChainConfig testInfo = new TestInfoWithChainConfig();
  public Address add;

  @BeforeEach
  public void init(TestInfo testInfo) {
      TracerTestBase.testInfo.chainConfig =
            switch (System.getProperty("unit.replay.tests.fork")) {
              case "LONDON" -> ChainConfig.MAINNET_TESTCONFIG(Fork.LONDON);
              case "PARIS" -> ChainConfig.MAINNET_TESTCONFIG(Fork.PARIS);
              case "SHANGHAI" -> ChainConfig.MAINNET_TESTCONFIG(Fork.SHANGHAI);
              case "CANCUN" -> ChainConfig.MAINNET_TESTCONFIG(Fork.CANCUN);
              case "PRAGUE" -> ChainConfig.MAINNET_TESTCONFIG(Fork.PRAGUE);
              default -> throw new IllegalArgumentException(
                  "Unknown fork: " + System.getProperty("unit.replay.tests.fork"));
            };
    TracerTestBase.testInfo.testInfo = testInfo;
  }
}
