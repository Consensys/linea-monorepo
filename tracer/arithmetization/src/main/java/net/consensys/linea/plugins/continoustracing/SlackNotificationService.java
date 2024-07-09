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

import static com.slack.api.model.block.Blocks.asBlocks;
import static com.slack.api.model.block.Blocks.divider;
import static com.slack.api.model.block.Blocks.section;
import static com.slack.api.model.block.composition.BlockCompositions.markdownText;

import java.io.IOException;
import java.util.stream.Collectors;

import com.slack.api.Slack;
import com.slack.api.webhook.Payload;
import com.slack.api.webhook.WebhookResponse;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import org.hyperledger.besu.datatypes.Hash;

@Slf4j
public class SlackNotificationService {
  final Slack slack;
  final String webHookUrl;

  protected SlackNotificationService(final Slack slack, final String webHookUrl) {
    this.slack = slack;
    this.webHookUrl = webHookUrl;
  }

  public static SlackNotificationService create(final String webHookUrl) {
    return new SlackNotificationService(Slack.getInstance(), webHookUrl);
  }

  public void sendCorsetFailureNotification(
      final long blockNumber, final String blockHash, final CorsetValidator.Result validationResult)
      throws IOException {
    final Payload messagePayload =
        Payload.builder()
            .text("Slack couldn't properly display the message.")
            .blocks(
                asBlocks(
                    section(
                        section ->
                            section.text(
                                markdownText(
                                    String.format(
                                        "*Trace verification failure for block %d (%s)",
                                        blockNumber, blockHash)))),
                    divider(),
                    section(
                        section ->
                            section.text(
                                markdownText(
                                    "Trace verification failed with the following error:\n\n"
                                        + "```"
                                        + validationResult
                                            .corsetOutput()
                                            // Remove all ANSI escape codes that Slack does not like
                                            .replaceAll("\u001B\\[[;\\d]*m", "")
                                        + "```\n\n"
                                        + "Trace file: "
                                        + validationResult.traceFile())))))
            .build();

    WebhookResponse response = slack.send(webHookUrl, messagePayload);
    checkResponse(response);
  }

  public void sendBlockTraceFailureNotification(
      final long blockNumber, final Hash txHash, final Throwable throwable) throws IOException {

    log.info("Throwable.getMessage(): {}", throwable.getMessage());

    final Payload messagePayload =
        Payload.builder()
            .text("Slack couldn't properly display the message.")
            .blocks(
                asBlocks(
                    section(
                        section ->
                            section.text(
                                markdownText(
                                    String.format(
                                        "*Error while tracing transaction %s in block %d*",
                                        txHash.toHexString(), blockNumber)))),
                    divider(),
                    section(
                        section ->
                            section.text(
                                markdownText(
                                    "Trace generation failed with the following error:\n\n"
                                        + "```"
                                        // more than 2000 characters will cause a 400 error when
                                        // sending the message
                                        + throwable
                                            .getMessage()
                                            .lines()
                                            .limit(3)
                                            .collect(Collectors.joining("\n"))
                                            .substring(0, 2000)
                                        + "```")))))
            .build();

    WebhookResponse response = slack.send(webHookUrl, messagePayload);
    checkResponse(response);
  }

  private void checkResponse(final WebhookResponse response) throws IOException {
    if (response.getCode() != 200) {
      throw new IOException(
          "Error while sending notification: status code: "
              + response.getCode()
              + ", body: "
              + response.getBody());
    }
  }
}
