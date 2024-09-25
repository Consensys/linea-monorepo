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

import static org.junit.jupiter.api.Assumptions.abort;

import java.time.Duration;
import java.util.List;

import ch.qos.logback.classic.Logger;
import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.read.ListAppender;
import lombok.Synchronized;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.jupiter.api.extension.TestWatcher;
import org.slf4j.LoggerFactory;

public class ReferenceTestWatcher implements TestWatcher {

  private static final int LOGBACK_POLL_ATTEMPTS = 100;
  private static final Duration LOGBACK_POLL_DELAY = Duration.ofMillis(10);

  public static final String JSON_INPUT_FILENAME = "failedBlockchainReferenceTests-input.json";
  public static final String JSON_OUTPUT_FILENAME = "failedBlockchainReferenceTests.json";
  ListAppender<ILoggingEvent> listAppender = new ListAppender<>();

  public ReferenceTestWatcher() {
    Logger logger = getLogbackLogger();
    listAppender.setContext(logger.getLoggerContext());
    listAppender.start();
    logger.addAppender(listAppender);
  }

  @Override
  public void testFailed(ExtensionContext context, Throwable cause) {
    String testName = context.getDisplayName().split(": ")[1];
    List<String> logEventMessages =
        listAppender.list.stream().map(ILoggingEvent::getMessage).toList();

    MapFailedReferenceTestsTool.mapAndStoreFailedReferenceTest(
        testName, logEventMessages, JSON_OUTPUT_FILENAME);
  }

  @Synchronized
  private static Logger getLogbackLogger() {
    try {
      org.slf4j.Logger slf4jLogger = null;
      for (int i = 0; i < LOGBACK_POLL_ATTEMPTS; i++) {
        slf4jLogger = LoggerFactory.getLogger(org.slf4j.Logger.ROOT_LOGGER_NAME);
        if (slf4jLogger instanceof ch.qos.logback.classic.Logger logbackLogger) {
          return logbackLogger;
        }
        Thread.sleep(LOGBACK_POLL_DELAY);
      }
      abort("SLF4J never returned a Logback logger. Last returned = " + slf4jLogger);
    } catch (InterruptedException ex) {
      abort("Thread interrupted while polling for Logback logger - " + ex);
    }
    throw new Error("unreachable code");
  }
}
