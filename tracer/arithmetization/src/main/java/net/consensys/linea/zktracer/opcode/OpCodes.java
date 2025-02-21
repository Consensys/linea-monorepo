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
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.type.CollectionType;
import com.fasterxml.jackson.databind.type.TypeFactory;
import lombok.SneakyThrows;
import net.consensys.linea.zktracer.json.JsonConverter;

/** Responsible for managing opcode loading and opcode metadata retrieval. */
public class OpCodes {
  private static final JsonConverter YAML_CONVERTER = JsonConverter.builder().enableYaml().build();

  private static final short OPCODES_LIST_SIZE = 256;
  public static List<OpCodeData> opCodeDataList = new ArrayList<>(OPCODES_LIST_SIZE);

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

    List<OpCodeData> opCodesLocal =
        YAML_CONVERTER.getObjectMapper().treeToValue(rootNode, typeReference);

    initOpcodes(opCodesLocal);
  }

  private static void initOpcodes(final List<OpCodeData> opCodesLocal) {
    for (int i = 0; i < OPCODES_LIST_SIZE; i++) {
      opCodeDataList.add(null);
    }
    for (OpCodeData opCodeData : opCodesLocal) {
      opCodeDataList.set(opCodeData.value(), opCodeData);
    }
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

    return Optional.ofNullable(opCodeDataList.get(value)).orElse(OpCodeData.forNonOpCodes(value));
  }

  /**
   * Get opcode metadata per opcode mnemonic of type {@link OpCode}.
   *
   * @param code opcode mnemonic of type {@link OpCode}.
   * @return an instance of {@link OpCodeData} corresponding to mnemonic of type {@link OpCode}.
   */
  public static OpCodeData of(final OpCode code) {
    return Optional.ofNullable(opCodeDataList.get(code.getOpcode()))
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
