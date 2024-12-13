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
package linea.plugin.acc.test.rpc.linea;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.util.List;

import linea.plugin.acc.test.LineaPluginTestBase;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import linea.plugin.acc.test.tests.web3j.generated.SimpleStorage;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.junit.jupiter.api.Test;

public class EstimateGasModuleLimitOverflowTest extends LineaPluginTestBase {
  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set(
            "--plugin-linea-module-limit-file-path=",
            getResourcePath("/txOverflowModuleLimits.toml"))
        .build();
  }

  @Test
  public void estimateGasFailsForExceedingModuleLineCountTest() throws Exception {

    final Account sender = accounts.getSecondaryBenefactor();

    final SimpleStorage simpleStorage = deploySimpleStorage();
    final String txData = simpleStorage.add(BigInteger.valueOf(100)).encodeFunctionCall();
    final var payload = Bytes.wrap(txData.getBytes(StandardCharsets.UTF_8));

    final EstimateGasTest.CallParams callParams =
        new EstimateGasTest.CallParams(
            null,
            sender.getAddress(),
            simpleStorage.getContractAddress(),
            null,
            payload.toHexString(),
            "0",
            null,
            null,
            null);

    final var reqLinea = new EstimateGasTest.BadLineaEstimateGasRequest(callParams);
    final var respLinea = reqLinea.execute(minerNode.nodeRequests());
    assertThat(respLinea.getCode()).isEqualTo(-32000);
    assertThat(respLinea.getMessage())
        .isEqualTo("Transaction line count for module SHF=31 is above the limit 20");
  }
}
