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

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.util.Preconditions.checkState;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.MalformedURLException;
import java.net.URISyntaxException;
import java.net.URL;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

/** Base class used for parsing JSON trace specs. */
@TestInstance(Lifecycle.PER_CLASS)
public abstract class AbstractModuleTracerBySpecTest extends AbstractBaseModuleTracerTest {
  private static final ObjectMapper MAPPER = new ObjectMapper();

  @ParameterizedTest(name = "{index} {0}")
  @MethodSource("specs")
  public void traceWithSpecFile(final String ignored, URL specUrl) throws IOException {
    traceOperation(specUrl);
  }

  private void traceOperation(final URL specFile) throws IOException {
    InputStream inputStream = specFile.openStream();
    final ObjectNode specNode = (ObjectNode) MAPPER.readTree(inputStream);
    final JsonNode request = specNode.get("input");
    final JsonNode expectedTrace = specNode.get("output");
    final JsonNode actualTrace = generateTrace(getModuleTracer().jsonKey(), request);

    assertThat(getTrace(actualTrace)).isEqualTo(getTrace(expectedTrace));
  }

  private static String getTrace(JsonNode actualTrace) throws JsonProcessingException {
    return MAPPER
        .writerWithDefaultPrettyPrinter()
        .writeValueAsString(MAPPER.readTree(String.valueOf(actualTrace)).findValue("Trace"));
  }

  private JsonNode generateTrace(String moduleName, JsonNode jsonNodeParams)
      throws JsonProcessingException {
    OpCode opcode = OpCode.valueOf(jsonNodeParams.get("opcode").asText());
    List<Bytes32> arguments = new ArrayList<>();
    JsonNode arg = jsonNodeParams.get("params");
    arg.forEach(bytes -> arguments.add(Bytes32.fromHexString(bytes.asText())));
    String trace = generateTrace(opcode, arguments);

    return MAPPER.readTree(trace).get(moduleName);
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
      final URL url = AbstractModuleTracerBySpecTest.class.getResource(path);
      checkState(url != null, "Cannot find test directory " + path);

      final Path dir;

      try {
        dir = Paths.get(url.toURI());
      } catch (final URISyntaxException e) {
        throw new RuntimeException("Problem converting URL to URI " + url, e);
      }

      try (final Stream<Path> s = Files.walk(dir, 1)) {
        s.map(Path::toFile)
            .filter(f -> f.getPath().endsWith(".json"))
            .map(AbstractModuleTracerBySpecTest::fileToParams)
            .forEach(specFiles::add);
      } catch (final IOException e) {
        throw new RuntimeException("Problem reading directory " + dir, e);
      }
    }

    final Object[][] result = new Object[specFiles.size()][2];
    for (int i = 0; i < specFiles.size(); i++) {
      result[i] = specFiles.get(i);
    }

    return result;
  }

  private static Object[] fileToParams(final File file) {
    try {
      final String fileName = file.toPath().getFileName().toString();
      final URL fileURL = file.toURI().toURL();

      return new Object[] {fileName, fileURL};
    } catch (final MalformedURLException e) {
      throw new RuntimeException("Problem reading spec file " + file.getAbsolutePath(), e);
    }
  }
}
