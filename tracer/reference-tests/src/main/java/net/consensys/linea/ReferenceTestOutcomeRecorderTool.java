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

import static net.consensys.linea.BlockchainReferenceTestJson.readBlockchainReferenceTestsOutput;
import static net.consensys.linea.BlockchainReferenceTestJson.writeToJsonFile;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import lombok.Synchronized;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.json.JsonConverter;

@Slf4j
public class ReferenceTestOutcomeRecorderTool {

  static JsonConverter jsonConverter = JsonConverter.builder().build();

  @Synchronized
  public static void mapAndStoreFailedReferenceTest(
      String testName, List<String> logEventMessages, String jsonOutputFilename) {
    Set<String> failedConstraints = getFailedConstraints(logEventMessages);
    if (!failedConstraints.isEmpty()) {
      readBlockchainReferenceTestsOutput(jsonOutputFilename)
          .thenCompose(
              jsonString -> {
                JsonConverter jsonConverter = JsonConverter.builder().build();

                BlockchainReferenceTestOutcome blockchainReferenceTestOutcome =
                    getBlockchainReferenceTestOutcome(jsonString);

                mapFailedConstraintsToTestsToModule(
                    blockchainReferenceTestOutcome.modulesToConstraints(),
                    failedConstraints,
                    testName);

                BlockchainReferenceTestOutcome finalBlockchainReferenceTestOutcome =
                    new BlockchainReferenceTestOutcome(
                        blockchainReferenceTestOutcome.failedCounter() + 1,
                        blockchainReferenceTestOutcome.successCounter(),
                        blockchainReferenceTestOutcome.modulesToConstraints());
                jsonString = jsonConverter.toJson(finalBlockchainReferenceTestOutcome);
                writeToJsonFile(jsonString, jsonOutputFilename);
                return null;
              });
    }
  }

  @Synchronized
  public static void incrementSuccessRate(String jsonOutputFilename) {
    readBlockchainReferenceTestsOutput(jsonOutputFilename)
        .thenAccept(
            jsonString -> {
              BlockchainReferenceTestOutcome blockchainReferenceTestOutcome =
                  getBlockchainReferenceTestOutcome(jsonString);

              BlockchainReferenceTestOutcome finalBlockchainReferenceTestOutcome =
                  new BlockchainReferenceTestOutcome(
                      blockchainReferenceTestOutcome.failedCounter(),
                      blockchainReferenceTestOutcome.successCounter() + 1,
                      blockchainReferenceTestOutcome.modulesToConstraints());
              jsonString = jsonConverter.toJson(finalBlockchainReferenceTestOutcome);
              writeToJsonFile(jsonString, jsonOutputFilename);
            });
  }

  @Synchronized
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
              testName, Objects.requireNonNull(moduleMapping), cleanedConstraintName);
      moduleMapping.constraints().put(cleanedConstraintName, failedTests);
    }
  }

  private static Set<String> aggregateFailedTestsForModuleConstraintPair(
      String testName, ModuleToConstraints moduleMapping, String cleanedConstraintName) {
    Set<String> failedTests =
        new HashSet<>(
            moduleMapping
                .constraints()
                .getOrDefault(cleanedConstraintName, Collections.emptySet()));
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

  public static ModuleToConstraints getModule(
      List<ModuleToConstraints> constraintToFailingTests, String moduleName) {
    List<ModuleToConstraints> modules =
        constraintToFailingTests.stream().filter(mapping -> mapping.equals(moduleName)).toList();

    if (modules.isEmpty()) {
      return null;
    } else {
      return modules.getFirst();
    }
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

  @Synchronized
  public static BlockchainReferenceTestOutcome getBlockchainReferenceTestOutcome(
      String jsonString) {
    BlockchainReferenceTestOutcome blockchainReferenceTestOutcome = null;
    if (!jsonString.isEmpty()) {
      blockchainReferenceTestOutcome =
          jsonConverter.fromJson(jsonString, BlockchainReferenceTestOutcome.class);
    } else {
      blockchainReferenceTestOutcome = new BlockchainReferenceTestOutcome(0, 0, new ArrayList<>());
    }
    return blockchainReferenceTestOutcome;
  }

  private static Set<String> getFailedConstraints(List<String> logEventMessages) {
    Set<String> failedConstraints = new HashSet<>();
    for (String eventMessage : logEventMessages) {
      failedConstraints.addAll(
          extractFailedConstraintsFromException(eventMessage.replaceAll("\u001B\\[[;\\d]*m", "")));
    }
    return failedConstraints;
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
    Pattern pattern = Pattern.compile("constraints failed: (.+)|failing constraint (.+)");
    Matcher matcher = pattern.matcher(message);

    List<String> exceptionCauses = new ArrayList<>();

    while (matcher.find()) {
      if (matcher.group(1) != null) {
        exceptionCauses.add(matcher.group(1));
      } else if (matcher.group(2) != null) {
        exceptionCauses.add(matcher.group(2));
      }
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
