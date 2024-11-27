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

import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ConcurrentSkipListSet;
import java.util.concurrent.atomic.AtomicInteger;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.json.JsonConverter;

@Slf4j
public class TestOutcomeWriterTool {

  public static void addFailed() {
    failedCounter.incrementAndGet();
  }

  public static JsonConverter jsonConverter = JsonConverter.builder().build();
  private static ObjectMapper objectMapper = new ObjectMapper();

  private static volatile AtomicInteger failedCounter = new AtomicInteger(0);
  private static volatile AtomicInteger successCounter = new AtomicInteger(0);
  private static volatile AtomicInteger disabledCounter = new AtomicInteger(0);
  private static volatile AtomicInteger abortedCounter = new AtomicInteger(0);
  private static volatile ConcurrentMap<
          String, ConcurrentMap<String, ConcurrentSkipListSet<String>>>
      modulesToConstraintsToTests = new ConcurrentHashMap<>(20);

  public static void addFailure(String type, String cause, String test) {
    failedCounter.incrementAndGet();
    modulesToConstraintsToTests
        .computeIfAbsent(type, t -> new ConcurrentHashMap<>())
        .computeIfAbsent(cause, t -> new ConcurrentSkipListSet<>())
        .add(test);
  }

  public static void addSuccess() {
    successCounter.incrementAndGet();
  }

  public static void addAborted() {
    abortedCounter.incrementAndGet();
  }

  public static void addSkipped() {
    disabledCounter.incrementAndGet();
  }

  public static void writeToJsonFile(String name) {
    String fileDirectory = getFileDirectory();
    log.info("writing results summary to {}", fileDirectory + "/" + name);
    try {
      Files.createDirectories(Path.of(fileDirectory));
    } catch (IOException e) {
      log.error("Error - Failed to create test directory output: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }
    try (FileWriter file = new FileWriter(Path.of(fileDirectory, name).toString())) {
      objectMapper.writeValue(
          file,
          new TestOutcome(
              failedCounter.get(),
              successCounter.get(),
              disabledCounter.get(),
              abortedCounter.get(),
              modulesToConstraintsToTests));
    } catch (Exception e) {
      log.error("Error - Failed to write test output: %s".formatted(e.getMessage()));
    }
  }

  public static String getFileDirectory() {
    String jsonDirectory = System.getenv("FAILED_TEST_JSON_DIRECTORY");
    if (jsonDirectory == null || jsonDirectory.isEmpty()) {
      return "../tmp/local/";
    }
    return jsonDirectory;
  }

  public static TestOutcome parseTestOutcome(String jsonString) {
    if (!jsonString.isEmpty()) {
      TestOutcome blockchainReferenceTestOutcome =
          jsonConverter.fromJson(jsonString, TestOutcome.class);
      return blockchainReferenceTestOutcome;
    }
    throw new RuntimeException("invalid JSON");
  }
}
