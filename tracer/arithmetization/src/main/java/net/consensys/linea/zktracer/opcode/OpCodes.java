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
import com.fasterxml.jackson.databind.module.SimpleModule;
import com.fasterxml.jackson.databind.type.CollectionType;
import com.fasterxml.jackson.databind.type.TypeFactory;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.Fork;
import net.consensys.linea.zktracer.json.JsonConverter;
import net.consensys.linea.zktracer.opcode.gas.Billing;
import net.consensys.linea.zktracer.opcode.gas.BillingDeserializer;

@Slf4j
/**
 * Responsible for managing opcode loading and opcode metadata retrieval. The design here is not
 * ideal for a concurrent setting, but it works. Essentially, the assumption is that loadOpcodes is
 * called before any attempt is made to access the opCodeDataList directly.
 */
public class OpCodes {
  private static final short OPCODES_LIST_SIZE = 256;
  private static volatile Fork forkLoaded;
  private static final List<OpCodeData> opCodeDataList = new ArrayList<>(OPCODES_LIST_SIZE);

  /** Loads all opcode metadata from src/main/resources/shanghaiOpcodes.yml. */
  @SneakyThrows(IOException.class)
  public static void loadOpcodes(Fork fork) {
    // Handle fast case where fork already loaded
    if (fork.equals(forkLoaded)) {
      log.info("opCodeDataList has already been initialized for " + fork + " fork.");
      return;
    } else if (forkLoaded != null) {
      throw new IllegalArgumentException(
          "request to load opcodes for "
              + fork
              + " conflicts with those previously "
              + "loaded for "
              + forkLoaded);
    }
    // Slow case.  Acquire lock and load opcodes.
    synchronized (opCodeDataList) {
      if (forkLoaded != null) {
        // Retry to get either an error or a warning above.
        loadOpcodes(fork);
        return;
      }
      // If we get here, then we are the only thread where forkLoaded == null.  Therefore, we can
      // proceed in peace.
      for (int i = 0; i < OPCODES_LIST_SIZE; i++) {
        opCodeDataList.add(null);
      }
      for (OpCodeData opCodeData : opCodesLocal(fork)) {
        opCodeDataList.set(opCodeData.value(), opCodeData);
      }
      // Finally, assign the fork.  Note, this must be done last.
      forkLoaded = fork;
    }
  }

  // Load opcodedata from yaml file.
  private static List<OpCodeData> opCodesLocal(Fork fork) throws IOException {
    final JsonConverter yamlConverter = JsonConverter.builder().enableYaml().build();

    final String yamlFileName = Fork.toString(fork) + "Opcodes.yml";

    final JsonNode rootNode =
        yamlConverter
            .getObjectMapper()
            .readTree(OpCodes.class.getClassLoader().getResourceAsStream(yamlFileName))
            .get("opcodes");

    final CollectionType typeReference =
        TypeFactory.defaultInstance().constructCollectionType(List.class, OpCodeData.class);

    SimpleModule module = new SimpleModule();
    switch (fork) {
      case LONDON, PARIS, SHANGHAI -> {
        // Before Cancun, we deserialize Billing with a type (TYPE_1, TYPE_2, TYPE_3, TYPE_4).
        module.addDeserializer(Billing.class, new BillingDeserializer(true));
      }
      case CANCUN, PRAGUE -> {
        // From Cancun and on, we deserialize Billing without a type.
        module.addDeserializer(Billing.class, new BillingDeserializer(false));
      }
      default -> throw new IllegalArgumentException("Unsupported fork: " + fork);
    }
    yamlConverter.getObjectMapper().registerModule(module);

    return yamlConverter.getObjectMapper().treeToValue(rootNode, typeReference);
  }

  /**
   * isValid checks whether or not a given opcode index corresponds with a real opcode.
   *
   * @param value
   * @return
   */
  public static boolean isValid(final int value) {
    if (value < 0 || value > 255) {
      throw new IllegalArgumentException("No OpCode with value %s is defined.".formatted(value));
    }
    synchronized (opCodeDataList) {
      if (opCodeDataList.isEmpty()) {
        throw new IllegalArgumentException("opcodes not initialised!");
      }
      return opCodeDataList.get(value) != null;
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
    synchronized (opCodeDataList) {
      if (opCodeDataList.isEmpty()) {
        throw new IllegalArgumentException("opcodes not initialised!");
      }
      //
      return Optional.ofNullable(opCodeDataList.get(value)).orElse(OpCodeData.forNonOpCodes(value));
    }
  }

  /**
   * Get opcode metadata per opcode mnemonic of type {@link OpCode}.
   *
   * @param code opcode mnemonic of type {@link OpCode}.
   * @return an instance of {@link OpCodeData} corresponding to mnemonic of type {@link OpCode}.
   */
  public static OpCodeData of(final OpCode code) {
    synchronized (opCodeDataList) {
      if (opCodeDataList.isEmpty()) {
        throw new IllegalArgumentException("opcodes not initialised!");
      }
      //
      return Optional.ofNullable(opCodeDataList.get(code.getOpcode()))
          .orElseThrow(
              () ->
                  new IllegalArgumentException(
                      "No OpCode of mnemonic %s is defined.".formatted(code)));
    }
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

  /**
   * Get an iterator over the underlying opcode data.
   *
   * @return
   */
  public static Iterable<OpCodeData> iterator() {
    synchronized (opCodeDataList) {
      if (opCodeDataList.isEmpty()) {
        throw new IllegalArgumentException("opcodes not initialised!");
      }
      //
      return opCodeDataList;
    }
  }
}
