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

package net.consensys.linea.zktracer;

import java.math.BigInteger;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonAnyGetter;
import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.SneakyThrows;
import net.consensys.linea.zktracer.json.BigIntegerSerializer;
import net.consensys.linea.zktracer.json.JsonConverter;
import net.consensys.linea.zktracer.json.NumericBooleanSerializer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@JsonPropertyOrder({"BlockRlp", "ParentRootHash", "TxNumber", "Pc", "Op", "shf", "shfRT"})
public record ZkTrace(
    @JsonProperty("BlockRlp") Bytes blockRlp,
    @JsonIgnore Bytes32 parentRootHash,
    @JsonProperty("TxNumber") BigInteger txNumber,
    @JsonProperty("Pc") BigInteger pc,
    @JsonProperty("Op") BigInteger op,
    @JsonAnyGetter Map<String, Object> traceResults) {

  private static final JsonConverter JSON_CONVERTER =
      JsonConverter.builder()
          .addCustomSerializer(Boolean.class, new NumericBooleanSerializer())
          .addCustomSerializer(BigInteger.class, new BigIntegerSerializer())
          .build();

  @JsonGetter("ParentRootHash")
  public String parentRootHashAsString() {
    return parentRootHash.toHexString();
  }

  /**
   * Converts traces to a JSON string.
   *
   * @return a trace as a JSON string
   */
  @SneakyThrows
  public String toJson() {
    return JSON_CONVERTER.toJson(
        new ZkTrace(
            null, Bytes32.ZERO, BigInteger.ZERO, BigInteger.ZERO, BigInteger.ZERO, traceResults));
  }
}
