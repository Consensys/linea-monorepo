/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.rpc.methods;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.concurrent.atomic.AtomicInteger;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.sequencer.forced.ForcedTransactionPoolService;
import net.consensys.linea.sequencer.forced.ForcedTransactionStatus;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonRpcParameter;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.RpcMethodError;

/**
 * RPC method for querying the inclusion status of a forced transaction.
 *
 * <p>Request format (same as eth_getTransactionReceipt):
 *
 * <pre>
 * {
 *   "method": "linea_getForcedTransactionInclusionStatus",
 *   "params": ["0xTRANSACTION_HASH"]
 * }
 * </pre>
 *
 * <p>Response format:
 *
 * <pre>
 * {
 *   "result": {
 *     "blockNumber": "0xeff35f",
 *     "from": "0x6221a9c005f6e47eb398fd867784cacfdcfff4e7",
 *     "inclusionResult": "BAD_NONCE",
 *     "transactionHash": "0xTRANSACTION_HASH"
 *   }
 * }
 * </pre>
 *
 * <p>Returns null if the transaction status is not yet known.
 */
@Slf4j
public class LineaGetForcedTransactionInclusionStatus {
  private static final AtomicInteger LOG_SEQUENCE = new AtomicInteger();

  private final JsonRpcParameter parameterParser = new JsonRpcParameter();
  private ForcedTransactionPoolService forcedTransactionPoolService;

  public LineaGetForcedTransactionInclusionStatus init(
      final ForcedTransactionPoolService forcedTransactionPoolService) {
    this.forcedTransactionPoolService = forcedTransactionPoolService;
    return this;
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "getForcedTransactionInclusionStatus";
  }

  public InclusionStatusResponse execute(final PluginRpcRequest request) {
    final int logId = log.isDebugEnabled() ? LOG_SEQUENCE.incrementAndGet() : -1;

    try {
      final String txHashHex = parseRequest(logId, request.getParams());
      final Hash txHash = Hash.fromHexString(txHashHex);

      log.atDebug()
          .setMessage("action=get_inclusion_status logId={} txHash={}")
          .addArgument(logId)
          .addArgument(txHashHex)
          .log();

      return forcedTransactionPoolService
          .getInclusionStatus(txHash)
          .map(
              status -> {
                log.atDebug()
                    .setMessage(
                        "action=get_inclusion_status_found logId={} txHash={} blockNumber={} result={}")
                    .addArgument(logId)
                    .addArgument(txHashHex)
                    .addArgument(status::blockNumber)
                    .addArgument(status::inclusionResult)
                    .log();
                return new InclusionStatusResponse(status);
              })
          .orElseGet(
              () -> {
                log.atDebug()
                    .setMessage("action=get_inclusion_status_not_found logId={} txHash={}")
                    .addArgument(logId)
                    .addArgument(txHashHex)
                    .log();
                return null;
              });

    } catch (final IllegalArgumentException e) {
      log.atWarn()
          .setMessage("action=get_inclusion_status_invalid_request logId={} error={}")
          .addArgument(logId)
          .addArgument(e::getMessage)
          .log();
      throw new PluginRpcEndpointException(new GetInclusionStatusError(e.getMessage()));
    } catch (final Exception e) {
      log.atError()
          .setMessage("action=get_inclusion_status_error logId={} error={}")
          .addArgument(logId)
          .addArgument(e::getMessage)
          .setCause(e)
          .log();
      throw new PluginRpcEndpointException(
          new GetInclusionStatusError("Internal error: " + e.getMessage()));
    }
  }

  private String parseRequest(final int logId, final Object[] params) {
    try {
      final String txHash = parameterParser.required(params, 0, String.class);
      if (txHash == null || txHash.isEmpty()) {
        throw new IllegalArgumentException("Transaction hash is required");
      }
      return txHash;
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

  /** Response class for inclusion status query. */
  public static class InclusionStatusResponse {
    @JsonProperty("blockNumber")
    public final String blockNumber;

    @JsonProperty("from")
    public final String from;

    @JsonProperty("inclusionResult")
    public final String inclusionResult;

    @JsonProperty("transactionHash")
    public final String transactionHash;

    public InclusionStatusResponse(final ForcedTransactionStatus status) {
      this.blockNumber = "0x" + Long.toHexString(status.blockNumber());
      this.from = status.from().toHexString();
      this.inclusionResult = status.inclusionResult().name();
      this.transactionHash = status.transactionHash().toHexString();
    }
  }

  static class GetInclusionStatusError implements RpcMethodError {
    private final String errMessage;

    GetInclusionStatusError(final String errMessage) {
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
