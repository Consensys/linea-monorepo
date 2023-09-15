/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.opcode.gas;

import java.util.Optional;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import lombok.SneakyThrows;

/** Custom Jackson deserializer for handling {@link Billing} properties. */
public class BillingDeserializer extends StdDeserializer<Billing> {

  public BillingDeserializer() {
    this(Billing.class);
  }

  protected BillingDeserializer(Class<?> vc) {
    super(vc);
  }

  @SneakyThrows
  @Override
  public Billing deserialize(JsonParser p, DeserializationContext context) {
    JsonNode node = p.getCodec().readTree(p);

    Optional<JsonNode> byWord = Optional.ofNullable(node.get("byWord"));
    Optional<JsonNode> byMxp = Optional.ofNullable(node.get("byMxp"));
    Optional<JsonNode> byByte = Optional.ofNullable(node.get("byByte"));

    if (byWord.isPresent()) {
      JsonNode wordNode = byWord.get();

      JsonNode wordPriceNode =
          Optional.of(wordNode.get("wordPrice"))
              .orElseThrow(
                  () ->
                      new IllegalArgumentException(
                          "'wordPrice' is a mandatory property when declaring 'byWord' billing"));

      MxpType type = extractMxpType(wordNode, "byWord");
      GasConstants wordPrice = GasConstants.valueOf(wordPriceNode.textValue());

      return Billing.byWord(type, wordPrice);
    }

    if (byMxp.isPresent()) {
      JsonNode mxpNode = byMxp.get();

      MxpType type = extractMxpType(mxpNode, "byMxp");

      return Billing.byMxp(type);
    }

    if (byByte.isPresent()) {
      JsonNode byteNode = byByte.get();

      JsonNode bytePriceNode =
          Optional.of(byteNode.get("bytePrice"))
              .orElseThrow(
                  () ->
                      new IllegalArgumentException(
                          "'bytePrice' is a mandatory property when declaring 'byByte' billing"));

      MxpType type = extractMxpType(byteNode, "byByte");
      GasConstants bytePrice = GasConstants.valueOf(bytePriceNode.textValue());

      return Billing.byByte(type, bytePrice);
    }

    return new Billing();
  }

  private MxpType extractMxpType(JsonNode node, String billingRate) {
    JsonNode typeNode =
        Optional.of(node.get("type"))
            .orElseThrow(
                () ->
                    new IllegalArgumentException(
                        "'mnemonic' is a mandatory property when declaring '%s' billing"
                            .formatted(billingRate)));

    return MxpType.valueOf(typeNode.textValue());
  }
}
