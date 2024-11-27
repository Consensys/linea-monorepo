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

import java.util.Optional;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.reporting.TestOutcomeWriterTool;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.jupiter.api.extension.TestWatcher;

@Slf4j
public class UnitTestWatcher implements TestWatcher {

  private String FAILED = "FAILED";

  @Override
  public void testFailed(ExtensionContext context, Throwable cause) {
    String testName = context.getDisplayName();
    log.info("Adding failure for {}", testName);
    TestOutcomeWriterTool.addFailure(
        FAILED, cause.getMessage().split(System.lineSeparator(), 2)[0], testName);
    log.info("Failure added for {}", testName);
  }

  @Override
  public void testSuccessful(ExtensionContext context) {
    TestOutcomeWriterTool.addSuccess();
  }

  @Override
  public void testDisabled(ExtensionContext context, Optional<String> reason) {
    TestOutcomeWriterTool.addSkipped();
  }

  @Override
  public void testAborted(ExtensionContext context, Throwable cause) {
    TestOutcomeWriterTool.addAborted();
  }
}
