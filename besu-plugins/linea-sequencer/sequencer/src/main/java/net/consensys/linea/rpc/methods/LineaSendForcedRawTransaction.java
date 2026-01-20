/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.rpc.methods;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.sequencer.forced.ForcedTransaction;
import net.consensys.linea.sequencer.forced.ForcedTransactionPoolService;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonRpcParameter;
import org.hyperledger.besu.ethereum.api.util.DomainObjectDecodeUtils;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.RpcMethodError;

/**
 * RPC method for submitting forced transactions. These transactions have highest priority in block
 * building and are guaranteed to be tried in order.
 *
 * <p>Request format:
 *
 * <pre>
 * {
 *   "method": "linea_sendForcedRawTransaction",
 *   "params": [
 *     {"transaction": "0x...", "deadline": "0x..."},
 *     {"transaction": "0x...", "deadline": "0x..."}
 *   ]
 * }
 * </pre>
 *
 * <p>Response format:
 *
 * <pre>
 * {
 *   "result": ["0xTX_HASH_1", "0xTX_HASH_2"]
 * }
 * </pre>
 */
@Slf4j
public class LineaSendForcedRawTransaction {
  private static final AtomicInteger LOG_SEQUENCE = new AtomicInteger();

  private final JsonRpcParameter parameterParser = new JsonRpcParameter();
  private ForcedTransactionPoolService forcedTransactionPoolService;

  public LineaSendForcedRawTransaction init(
      final ForcedTransactionPoolService forcedTransactionPoolService) {
    this.forcedTransactionPoolService = forcedTransactionPoolService;
    return this;
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "sendForcedRawTransaction";
  }

  public List<String> execute(final PluginRpcRequest request) {
    final int logId = log.isDebugEnabled() ? LOG_SEQUENCE.incrementAndGet() : -1;

    try {
      final ForcedTransactionParam[] params = parseRequest(logId, request.getParams());

      if (params == null || params.length == 0) {
        throw new IllegalArgumentException("At least one forced transaction is required");
      }

      log.atDebug()
          .setMessage("action=send_forced_raw_tx_received logId={} count={}")
          .addArgument(logId)
          .addArgument(params.length)
          .log();

      final List<ForcedTransaction> forcedTransactions = new ArrayList<>(params.length);

      for (int i = 0; i < params.length; i++) {
        final ForcedTransactionParam param = params[i];

        if (param.transaction == null || param.transaction.isEmpty()) {
          throw new IllegalArgumentException(
              "Transaction at index " + i + " has empty transaction data");
        }

        final Transaction tx = decodeTransaction(param.transaction, i);
        final long deadline = parseDeadline(param.deadline, i);

        forcedTransactions.add(new ForcedTransaction(tx.getHash(), tx, deadline));

        log.atDebug()
            .setMessage("action=parse_forced_tx logId={} index={} txHash={} deadline={}")
            .addArgument(logId)
            .addArgument(i)
            .addArgument(tx.getHash()::toHexString)
            .addArgument(deadline)
            .log();
      }

      final List<Hash> hashes =
          forcedTransactionPoolService.addForcedTransactions(forcedTransactions);

      log.atInfo()
          .setMessage("action=send_forced_raw_tx_success logId={} count={}")
          .addArgument(logId)
          .addArgument(hashes::size)
          .log();

      return hashes.stream().map(Hash::toHexString).toList();

    } catch (final IllegalArgumentException e) {
      log.atWarn()
          .setMessage("action=send_forced_raw_tx_invalid logId={} error={}")
          .addArgument(logId)
          .addArgument(e::getMessage)
          .log();
      throw new PluginRpcEndpointException(new SendForcedRawTransactionError(e.getMessage()));
    } catch (final Exception e) {
      log.atError()
          .setMessage("action=send_forced_raw_tx_error logId={} error={}")
          .addArgument(logId)
          .addArgument(e::getMessage)
          .setCause(e)
          .log();
      throw new PluginRpcEndpointException(
          new SendForcedRawTransactionError("Internal error: " + e.getMessage()));
    }
  }

  private ForcedTransactionParam[] parseRequest(final int logId, final Object[] params) {
    try {
      return parameterParser.required(params, 0, ForcedTransactionParam[].class);
    } catch (final Exception e) {
      log.atError()
          .setMessage("action=parse_request_failed logId={} error={}")
          .addArgument(logId)
          .addArgument(e::getMessage)
          .setCause(e)
          .log();
      throw new IllegalArgumentException("Malformed request parameters: " + e.getMessage());
    }
  }

  private Transaction decodeTransaction(final String rawTx, final int index) {
    try {
      return DomainObjectDecodeUtils.decodeRawTransaction(rawTx);
    } catch (final Exception e) {
      throw new IllegalArgumentException(
          "Failed to decode transaction at index " + index + ": " + e.getMessage());
    }
  }

  private long parseDeadline(final String deadlineHex, final int index) {
    if (deadlineHex == null || deadlineHex.isEmpty()) {
      throw new IllegalArgumentException("Deadline is required at index " + index);
    }
    try {
      return Long.decode(deadlineHex);
    } catch (final NumberFormatException e) {
      throw new IllegalArgumentException(
          "Invalid deadline format at index " + index + ": " + deadlineHex);
    }
  }

  /** Parameter class for forced transaction submission. */
  public static class ForcedTransactionParam {
    @JsonProperty("transaction")
    public String transaction;

    @JsonProperty("deadline")
    public String deadline;

    @JsonCreator
    public ForcedTransactionParam(
        @JsonProperty("transaction") final String transaction,
        @JsonProperty("deadline") final String deadline) {
      this.transaction = transaction;
      this.deadline = deadline;
    }
  }

  static class SendForcedRawTransactionError implements RpcMethodError {
    private final String errMessage;

    SendForcedRawTransactionError(final String errMessage) {
      this.errMessage = errMessage;
    }

    @Override
    public int getCode() {
      return INVALID_PARAMS_ERROR_CODE;
    }

    @Override
    public String getMessage() {
      return errMessage;
    }
  }
}
