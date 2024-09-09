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
import java.nio.file.FileAlreadyExistsException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

import lombok.Synchronized;
import lombok.extern.slf4j.Slf4j;

@Slf4j
public class FailedTestJson {
  static String fileDirectory = System.getProperty("FAILED_TEST_JSON_DIRECTORY", "../tmp/local/");

  @Synchronized
  public static String readFailedTestsOutput(String fileName) {
    Path directoryPath = Paths.get(fileDirectory);
    Path filePath = Paths.get(fileDirectory + fileName);
    String jsonString = "";

    try {
      jsonString = new String(Files.readAllBytes(filePath));
    } catch (IOException e) {
      log.debug(
          "Failed to read json output, could be first time running: %s".formatted(e.getMessage()));

      try {
        Files.createDirectories(directoryPath);
      } catch (FileAlreadyExistsException x) {
        log.debug("Directory %s already exists.".formatted(directoryPath));
      } catch (IOException ex) {
        log.error("Error - Failed to create directory: %s".formatted(e));
      }

      try {
        Files.createFile(filePath);
        log.debug("Created a new file at: %s".formatted(filePath));
      } catch (IOException ex) {
        log.error("Failed to create a new file at: %s".formatted(filePath), ex);
      }
    }
    return jsonString;
  }

  @Synchronized
  public static void writeToJsonFile(String jsonString, String fileName) {
    try (FileWriter file = new FileWriter(fileDirectory + fileName)) {
      file.write(jsonString);
    } catch (Exception e) {
      log.error("Error - Failed to write failed test output: %s".formatted(e.getMessage()));
    }
  }
}
