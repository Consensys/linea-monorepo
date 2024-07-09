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
package net.consensys.linea.plugins.continoustracing;

import java.io.IOException;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.plugins.exception.InvalidTraceHandlerException;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.plugin.data.BlockHeader;

@Slf4j
public class TraceFailureHandler {
  final SlackNotificationService slackNotificationService;

  public TraceFailureHandler(final SlackNotificationService slackNotificationService) {
    this.slackNotificationService = slackNotificationService;
  }

  public void handleCorsetFailure(
      final BlockHeader blockHeader, final CorsetValidator.Result result)
      throws InvalidTraceHandlerException {
    try {
      slackNotificationService.sendCorsetFailureNotification(
          blockHeader.getNumber(), blockHeader.getBlockHash().toHexString(), result);
    } catch (IOException e) {
      log.error("Error while sending slack notification: {}", e.getMessage());
      throw new InvalidTraceHandlerException(e);
    }
  }

  public void handleBlockTraceFailure(
      final long blockNumber, final Hash txHash, final Throwable throwable) {
    try {
      slackNotificationService.sendBlockTraceFailureNotification(blockNumber, txHash, throwable);
    } catch (IOException e) {
      log.error("Error while handling block trace failure: {}", e.getMessage());
    }
  }
}
