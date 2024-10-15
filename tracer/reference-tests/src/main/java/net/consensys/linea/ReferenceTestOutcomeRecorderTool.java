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
package net.consensys.linea;

import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ConcurrentSkipListSet;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.Synchronized;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.json.JsonConverter;

@Slf4j
public class ReferenceTestOutcomeRecorderTool {

  public static final String JSON_INPUT_FILENAME = "failedBlockchainReferenceTests-input.json";
  public static final String JSON_OUTPUT_FILENAME = "failedBlockchainReferenceTests.json";
  public static JsonConverter jsonConverter = JsonConverter.builder().build();
  private static volatile AtomicInteger failedCounter = new AtomicInteger(0);
  private static volatile AtomicInteger successCounter = new AtomicInteger(0);
  private static volatile AtomicInteger disabledCounter = new AtomicInteger(0);
  private static volatile AtomicInteger abortedCounter = new AtomicInteger(0);
  private static volatile ConcurrentMap<
          String, ConcurrentMap<String, ConcurrentSkipListSet<String>>>
      modulesToConstraintsToTests = new ConcurrentHashMap<>();

  public static void mapAndStoreTestResult(
      String testName, TestState success, Map<String, Set<String>> failedConstraints) {
    switch (success) {
      case FAILED -> {
        failedCounter.incrementAndGet();
        for (Map.Entry<String, Set<String>> failedConstraint : failedConstraints.entrySet()) {
          String moduleName = failedConstraint.getKey();
          for (String constraint : failedConstraint.getValue()) {
            ConcurrentMap<String, ConcurrentSkipListSet<String>> constraintsToTests =
                modulesToConstraintsToTests.computeIfAbsent(
                    moduleName, m -> new ConcurrentHashMap<>());
            ConcurrentSkipListSet<String> failingTests =
                constraintsToTests.computeIfAbsent(constraint, m -> new ConcurrentSkipListSet<>());
            int size = failingTests.size();
            failingTests.add(testName);
            if (failingTests.size() == size) {
              log.warn("Duplicate name found... {}", failedConstraint);
            }
          }
        }
      }
      case SUCCESS -> successCounter.incrementAndGet();
      case ABORTED -> abortedCounter.incrementAndGet();
      case DISABLED -> disabledCounter.incrementAndGet();
    }
  }

  @Synchronized
  public static BlockchainReferenceTestOutcome parseBlockchainReferenceTestOutcome(
      String jsonString) {
    if (!jsonString.isEmpty()) {
      BlockchainReferenceTestOutcome blockchainReferenceTestOutcome =
          jsonConverter.fromJson(jsonString, BlockchainReferenceTestOutcome.class);
      return blockchainReferenceTestOutcome;
    }
    throw new RuntimeException("invalid JSON");
  }

  public static Map<String, Set<String>> extractConstraints(String message) {
    Map<String, Set<String>> pairs = new HashMap<>();
    String cleaned = message.replaceAll(("\\[[0-9]+m"), " ").replace(']', ' ').trim();

    // case where corset sends constraint failed and the list of constraints
    if (message.contains("constraints failed:")) {
      cleaned =
          cleaned.substring(
              message.indexOf("constraints failed:") + "constraints failed:".length());
      String[] constraints = cleaned.split(",");
      for (int i = 0; i < constraints.length; i++) {
        getPairFromString(constraints[i], pairs);
      }
    } else if (message.contains("failing constraint")) {
      // case where corset sends failing constraint with constraints one by one
      String[] lines = cleaned.split("\\n");
      for (int i = 0; i < lines.length; i++) {
        if (lines[i].contains("failing constraint")) {
          String line =
              lines[i].substring(
                  lines[i].indexOf("failing constraint") + "failing constraint".length());
          line = line.replace(':', ' ');
          getPairFromString(line, pairs);
        }
      }
    } else {
      // case where corset can't expend the trace
      if (message.contains("Error: while expanding ")) {
        String[] lines = cleaned.split("\\n");
        for (int i = 0; i < lines.length; i++) {
          if (lines[i].contains("reading data for")) {
            String regex = "for\\s+(\\w+)\\s+\\.\\s+(\\w+)";
            Pattern pattern = Pattern.compile(regex);
            Matcher matcher = pattern.matcher(lines[i]);
            if (matcher.find()) {
              String module = matcher.group(1);
              String constraint = matcher.group(2);
              pairs
                  .computeIfAbsent("Expanding " + module.trim(), p -> new HashSet<>())
                  .add(constraint.trim());
            } else {
              pairs.computeIfAbsent("UNKNOWN", p -> new HashSet<>()).add("UNKNOWN");
            }
          }
        }
      } else if (message.contains("lookup ")) {
        String regex = "\"lookup \"(.*)\" failed";
        Pattern pattern = Pattern.compile(regex);
        Matcher matcher = pattern.matcher(message);
        if (matcher.find()) {
          String module = "LOOKUP";
          String constraint = matcher.group(1);
          pairs.computeIfAbsent(module.trim(), p -> new HashSet<>()).add(constraint.trim());
        }
      } else if (message.contains("out-of-bounds")) {
        String regex = "column (.*) is out-of-bounds";
        Pattern pattern = Pattern.compile(regex);
        Matcher matcher = pattern.matcher(message);
        if (matcher.find()) {
          String[] group = matcher.group(1).split("\\.");
          String module = group[0];
          String constraint = group[1];
          pairs.computeIfAbsent(module.trim(), p -> new HashSet<>()).add(constraint.trim());
        }
      } else {
        log.info("can't extract constraint, setting UNKNOWN for {}", message);
        pairs.computeIfAbsent("UNKNOWN", p -> new HashSet<>()).add("UNKNOWN");
      }
    }
    return pairs;
  }

  private static void getPairFromString(String constraint, Map<String, Set<String>> pairs) {
    String[] pair;
    if (constraint.contains("-into-")) {
      pair = constraint.split("-into-");
    } else {
      pair = constraint.split("\\.");
    }
    pairs.computeIfAbsent(pair[0].trim(), p -> new HashSet<>()).add(pair[1].trim());
  }

  @Synchronized
  public static void writeToJsonFile(String name) {
    try {
      String directory = setFileDirectory();
      log.info("Reference test will be written to file {} \\ {}", directory, JSON_OUTPUT_FILENAME);
      writeToJsonFileInternal(name).get();
      log.info("Reference test results written to file {}", JSON_OUTPUT_FILENAME);
      log.info(
          "Path exists: {}, file exist: {}",
          Paths.get(directory).toFile().exists(),
          Paths.get(directory).resolve(JSON_OUTPUT_FILENAME).toFile().exists());
    } catch (Exception e) {
      log.error("Error while writing results");
      throw new RuntimeException("Error while writing results", e);
    }
  }

  static ObjectMapper objectMapper = new ObjectMapper();

  @Synchronized
  private static CompletableFuture<Void> writeToJsonFileInternal(String name) {
    String fileDirectory = setFileDirectory();
    log.info("writing results summary to {}", fileDirectory + "/" + name);
    try {
      Files.createDirectories(Path.of(fileDirectory));
    } catch (IOException e) {
      log.error("Error - Failed to create test directory output: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }
    return CompletableFuture.runAsync(
        () -> {
          try (FileWriter file = new FileWriter(fileDirectory + name)) {
            objectMapper.writeValue(
                file,
                new BlockchainReferenceTestOutcome(
                    failedCounter.get(),
                    successCounter.get(),
                    disabledCounter.get(),
                    abortedCounter.get(),
                    modulesToConstraintsToTests));
          } catch (Exception e) {
            log.error("Error - Failed to write test output: %s".formatted(e.getMessage()));
          }
        });
  }

  static String setFileDirectory() {
    String jsonDirectory = System.getenv("FAILED_TEST_JSON_DIRECTORY");
    if (jsonDirectory == null || jsonDirectory.isEmpty()) {
      return "../tmp/local/";
    }
    return jsonDirectory;
  }
}
