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

package net.consensys.linea.zktracer.testing;

import static org.assertj.core.api.Assertions.assertThat;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.MalformedURLException;
import java.net.URL;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import lombok.SneakyThrows;
import net.consensys.linea.zktracer.json.JsonConverter;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;

/** Base class used for parsing JSON trace specs. */
public class SpecTests {
  private static final JsonConverter JSON_CONVERTER =
      JsonConverter.builder().enablePrettyPrint().build();

  /**
   * Runs a test that parses a JSON spec, specifying trace params and compares then against the
   * actual generated trace.
   *
   * @param specFile spec file containing the expected trace params
   * @param moduleName name of the module being tested
   */
  @SneakyThrows(IOException.class)
  public static void runSpecTestWithTraceComparison(final URL specFile, final String moduleName) {
    InputStream inputStream = specFile.openStream();
    ObjectNode specNode = (ObjectNode) JSON_CONVERTER.getObjectMapper().readTree(inputStream);
    JsonNode request = specNode.get("input");
    JsonNode expectedTrace = specNode.get("output");
    JsonNode actualTrace = generateTraceFromSpecParams(moduleName, request);

    assertThat(getTrace(actualTrace)).isEqualTo(getTrace(expectedTrace));
  }

  /**
   * Find trace spec JSON file paths.
   *
   * @param subDirectoryPaths directories with trace spec JSON file paths.
   * @return trace spec JSON file paths
   */
  public static Object[][] findSpecFiles(final String... subDirectoryPaths) {
    final List<Object[]> specFiles = new ArrayList<>();

    for (final String path : subDirectoryPaths) {
      Path specDir = Paths.get("src/test/resources/specs/%s".formatted(path)).toAbsolutePath();

      try (final Stream<Path> s = Files.walk(specDir, 1)) {
        s.map(Path::toFile)
            .filter(f -> f.getPath().endsWith(".json"))
            .map(SpecTests::fileToParams)
            .forEach(specFiles::add);
      } catch (final IOException e) {
        throw new RuntimeException("Problem reading directory " + specDir, e);
      }
    }

    final Object[][] result = new Object[specFiles.size()][2];
    for (int i = 0; i < specFiles.size(); i++) {
      result[i] = specFiles.get(i);
    }

    return result;
  }

  private static String getTrace(JsonNode actualTrace) throws JsonProcessingException {
    JsonNode traceJsonNode =
        JSON_CONVERTER.getObjectMapper().readTree(String.valueOf(actualTrace)).findValue("Trace");

    return JSON_CONVERTER.toJson(traceJsonNode);
  }

  private static JsonNode generateTraceFromSpecParams(String moduleName, JsonNode jsonNodeParams)
      throws JsonProcessingException {
    OpCode opCode = OpCode.valueOf(jsonNodeParams.get("opcode").asText());
    List<Bytes32> arguments = new ArrayList<>();
    JsonNode arg = jsonNodeParams.get("params");
    arg.forEach(bytes -> arguments.add(0, Bytes32.fromHexString(bytes.asText())));
    String trace = ModuleTests.generateTrace(opCode, arguments);

    return JSON_CONVERTER.getObjectMapper().readTree(trace).get(moduleName);
  }

  private static Object[] fileToParams(final File file) {
    try {
      final String fileName = file.toPath().getFileName().toString();
      final URL fileURL = file.toURI().toURL();

      return new Object[] {fileName, fileURL};
    } catch (final MalformedURLException e) {
      throw new RuntimeException(
          "Problem reading spec file %s: %s".formatted(file.getAbsolutePath(), e));
    }
  }
}
