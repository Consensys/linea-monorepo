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

package net.consensys.linea.zktracer.opcode;

import java.io.IOException;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.type.CollectionType;
import com.fasterxml.jackson.databind.type.TypeFactory;
import lombok.SneakyThrows;
import net.consensys.linea.zktracer.json.JsonConverter;

/** Responsible for managing opcode loading and opcode metadata retrieval. */
public class OpCodes {
  private static final JsonConverter YAML_CONVERTER = JsonConverter.builder().enableYaml().build();

  static Map<Integer, OpCodeData> valueToOpCodeDataMap;
  public static Map<OpCode, OpCodeData> opCodeToOpCodeDataMap;

  static {
    init();
  }

  /** Loads all opcode metadata from src/main/resources/opcodes.yml. */
  @SneakyThrows(IOException.class)
  private static void init() {
    JsonNode rootNode =
        YAML_CONVERTER
            .getObjectMapper()
            .readTree(OpCodes.class.getClassLoader().getResourceAsStream("opcodes.yml"))
            .get("opcodes");

    CollectionType typeReference =
        TypeFactory.defaultInstance().constructCollectionType(List.class, OpCodeData.class);

    List<OpCodeData> opCodes =
        YAML_CONVERTER.getObjectMapper().treeToValue(rootNode, typeReference);

    valueToOpCodeDataMap = opCodes.stream().collect(Collectors.toMap(OpCodeData::value, e -> e));
    opCodeToOpCodeDataMap =
        opCodes.stream().collect(Collectors.toMap(OpCodeData::mnemonic, e -> e));
  }

  /**
   * Get opcode metadata per opcode long value.
   *
   * @param value opcode value.
   * @return an instance of {@link OpCodeData} corresponding to the numeric value.
   */
  public static OpCodeData of(final int value) {
    if (value < 0 || value > 255) {
      throw new IllegalArgumentException("No OpCode with value %s is defined.".formatted(value));
    }

    return valueToOpCodeDataMap.getOrDefault(value, OpCodeData.forNonOpCodes(value));
  }

  /**
   * Get opcode metadata per opcode mnemonic of type {@link OpCode}.
   *
   * @param code opcode mnemonic of type {@link OpCode}.
   * @return an instance of {@link OpCodeData} corresponding to mnemonic of type {@link OpCode}.
   */
  public static OpCodeData of(final OpCode code) {
    return Optional.ofNullable(opCodeToOpCodeDataMap.get(code))
        .orElseThrow(
            () ->
                new IllegalArgumentException(
                    "No OpCode of mnemonic %s is defined.".formatted(code)));
  }

  /**
   * Get opcode metadata for a list of {@link OpCode}s.
   *
   * @param codes a list of opcode mnemonics of type {@link OpCode}.
   * @return a list of {@link OpCodeData} items corresponding their mnemonics of type {@link
   *     OpCode}.
   */
  public static List<OpCodeData> of(final OpCode... codes) {
    return Arrays.stream(codes).map(OpCodes::of).toList();
  }
}
