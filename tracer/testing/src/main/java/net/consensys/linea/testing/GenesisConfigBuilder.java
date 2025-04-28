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
package net.consensys.linea.testing;

import java.math.BigInteger;

import com.fasterxml.jackson.databind.node.ObjectNode;
import org.hyperledger.besu.config.GenesisAccount;
import org.hyperledger.besu.config.GenesisConfig;
import org.hyperledger.besu.config.JsonUtil;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;

public class GenesisConfigBuilder {
  private static ObjectNode LINEA_CONFIG =
      JsonUtil.objectNodeFromURL(GenesisConfigBuilder.class.getResource("/lineabesu.json"), true);

  private ObjectNode genesisRoot;
  private ObjectNode configNode;
  private ObjectNode allocNode;

  public GenesisConfigBuilder() {
    genesisRoot = LINEA_CONFIG.deepCopy();
    configNode = genesisRoot.withObjectProperty("config");
    allocNode = genesisRoot.withObjectProperty("alloc");
  }

  public GenesisConfigBuilder setChainId(BigInteger chainId) {
    configNode.put("chainId", chainId);
    return this;
  }

  public GenesisConfigBuilder addAccount(ToyAccount toyAccount) {
    GenesisAccount genesisAccount = toyAccount.toGenesisAccount();

    final ObjectNode accountNode =
        allocNode.withObjectProperty(genesisAccount.address().toHexString());
    accountNode.put("nonce", genesisAccount.nonce());
    accountNode.put("balance", genesisAccount.balance().toHexString());
    accountNode.put("code", genesisAccount.code().toHexString());

    ObjectNode accountStorageNode = accountNode.withObject("storage");
    genesisAccount
        .storage()
        .forEach(
            (key, value) -> {
              accountStorageNode.put(key.toHexString(), value.toHexString());
            });

    if (genesisAccount.privateKey() != null) {
      accountStorageNode.put("privatekey", genesisAccount.privateKey().toHexString());
    }

    return this;
  }

  public GenesisConfigBuilder setDifficulty(String difficulty) {
    genesisRoot.put("difficulty", difficulty);
    return this;
  }

  public GenesisConfigBuilder setGasLimit(long gasLimit) {
    genesisRoot.put("gasLimit", gasLimit);
    return this;
  }

  public GenesisConfigBuilder setBaseFeePerGas(Wei baseFeePerGas) {
    genesisRoot.put("baseFeePerGas", baseFeePerGas.toHexString());
    return this;
  }

  public GenesisConfigBuilder setCoinbase(Address coinbase) {
    genesisRoot.put("coinbase", coinbase.toHexString());
    return this;
  }

  public GenesisConfig build() {
    return GenesisConfig.fromConfig(genesisRoot);
  }

  public String buildAsString() {
    return genesisRoot.toPrettyString();
  }
}
