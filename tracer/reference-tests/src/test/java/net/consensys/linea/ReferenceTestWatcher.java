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

import java.time.LocalDate;
import java.util.List;

import ch.qos.logback.classic.Logger;
import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.read.ListAppender;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.jupiter.api.extension.TestWatcher;

public class ReferenceTestWatcher implements TestWatcher {

  public static final String JSON_OUTPUT_FILENAME =
      "failedBlockchainReferenceTests-%s.json".formatted(LocalDate.now().toString());
  ListAppender<ILoggingEvent> listAppender = new ListAppender<>();

  public ReferenceTestWatcher() {
    Logger logger = (Logger) org.slf4j.LoggerFactory.getLogger(org.slf4j.Logger.ROOT_LOGGER_NAME);
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
}
