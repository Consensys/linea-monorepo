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
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import lombok.Synchronized;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.json.JsonConverter;
import org.jetbrains.annotations.NotNull;

@Slf4j
public class MapFailedReferenceTestsTool {

  public static void mapAndStoreFailedReferenceTest(
      String testName, List<String> logEventMessages, String jsonOutputPath) {
    Set<String> failedConstraints = getFailedConstraints(logEventMessages);
    String jsonString = readFailedTestsOutput(jsonOutputPath);
    JsonConverter jsonConverter = JsonConverter.builder().build();

    List<ModuleToConstraints> modulesToConstraints =
        getModulesToConstraints(jsonString, jsonConverter);

    mapFailedConstraintsToTestsToModule(modulesToConstraints, failedConstraints, testName);

    jsonString = jsonConverter.toJson(modulesToConstraints);
    writeToJsonFile(jsonString, jsonOutputPath);
  }

  private static void mapFailedConstraintsToTestsToModule(
      List<ModuleToConstraints> modulesToConstraints,
      Set<String> failedConstraints,
      String testName) {
    for (String constraint : failedConstraints) {
      String moduleName = getModuleFromFailedConstraint(constraint);

      addModuleIfAbsent(modulesToConstraints, moduleName);
      ModuleToConstraints moduleMapping = getModule(modulesToConstraints, moduleName);
      String cleanedConstraintName = getCleanedConstraintName(constraint);

      Set<String> failedTests =
          aggregateFailedTestsForModuleConstraintPair(
              testName, moduleMapping, cleanedConstraintName);
      moduleMapping.constraints().put(cleanedConstraintName, failedTests.stream().toList());
    }
  }

  private static @NotNull Set<String> aggregateFailedTestsForModuleConstraintPair(
      String testName, ModuleToConstraints moduleMapping, String cleanedConstraintName) {
    Set<String> failedTests =
        new HashSet<>(
            moduleMapping
                .constraints()
                .getOrDefault(cleanedConstraintName, Collections.emptyList()));
    failedTests.add(testName);
    return failedTests;
  }

  private static String getCleanedConstraintName(String constraint) {
    String cleanedConstraintName = constraint;
    if (cleanedConstraintName.contains(".")) {
      cleanedConstraintName = cleanedConstraintName.split("\\.")[1];
    }
    return cleanedConstraintName;
  }

  private static ModuleToConstraints getModule(
      List<ModuleToConstraints> constraintToFailingTests, String moduleName) {
    return constraintToFailingTests.stream()
        .filter(mapping -> mapping.equals(moduleName))
        .toList()
        .getFirst();
  }

  private static void addModuleIfAbsent(
      List<ModuleToConstraints> constraintToFailingTests, String moduleName) {
    if (constraintToFailingTests.stream()
        .filter(mapping -> mapping.equals(moduleName))
        .toList()
        .isEmpty()) {
      constraintToFailingTests.add(new ModuleToConstraints(moduleName, new HashMap<>()));
    }
  }

  public static List<ModuleToConstraints> getModulesToConstraints(
      String jsonString, JsonConverter jsonConverter) {
    List<ModuleToConstraints> moduleToConstraints = new ArrayList<>();
    if (!jsonString.isEmpty()) {
      moduleToConstraints =
          new ArrayList<>(
              Arrays.asList(jsonConverter.fromJson(jsonString, ModuleToConstraints[].class)));
    }
    return moduleToConstraints;
  }

  private static Set<String> getFailedConstraints(List<String> logEventMessages) {
    Set<String> failedConstraints = new HashSet<>();
    for (String eventMessage : logEventMessages) {
      failedConstraints.addAll(
          extractFailedConstraintsFromException(eventMessage.replaceAll("\u001B\\[[;\\d]*m", "")));
    }
    return failedConstraints;
  }

  @Synchronized
  public static String readFailedTestsOutput(String filePath) {
    Path path = Paths.get(filePath);
    String jsonString = "";
    try {
      jsonString = new String(Files.readAllBytes(path));
    } catch (IOException e) {
      log.info(
          "Failed to read json output, could be first time running: %s".formatted(e.getMessage()));
      try {
        Files.createFile(path);
        log.info("Created a new file at: %s".formatted(filePath));
      } catch (IOException ex) {
        log.error("Failed to create a new file at: %s".formatted(filePath), ex);
      }
    }
    return jsonString;
  }

  @Synchronized
  public static void writeToJsonFile(String jsonString, String filePath) {
    try (FileWriter file = new FileWriter(filePath)) {
      file.write(jsonString);
    } catch (Exception e) {
      log.error("Error - Failed to write failed test output: %s".formatted(e.getMessage()));
    }
  }

  private static List<String> extractFailedConstraintsFromException(String message) {
    List<String> exceptionCauses = extractionExceptionCauses(message);
    Set<String> failedConstraints = new HashSet<>();
    for (String causes : exceptionCauses) {
      failedConstraints.addAll(extractSeparateConstraints(causes));
    }
    return failedConstraints.stream().toList();
  }

  private static List<String> extractionExceptionCauses(String message) {
    Pattern pattern = Pattern.compile("constraints failed: (.+)");
    Matcher matcher = pattern.matcher(message);

    List<String> exceptionCauses = new ArrayList<>();

    while (matcher.find()) {
      String cause = matcher.group(1);
      exceptionCauses.add(cause);
    }

    return exceptionCauses;
  }

  private static List<String> extractSeparateConstraints(String exceptionCauses) {
    String[] parts = exceptionCauses.split(", ");
    return new ArrayList<>(Arrays.asList(parts));
  }

  private static String getModuleFromFailedConstraint(String failedConstraint) {
    if (failedConstraint.contains(".")) {
      return failedConstraint.split("\\.")[0];
    } else {
      return failedConstraint.split("-")[0];
    }
  }
}
