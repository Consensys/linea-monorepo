package net.consensys.zkevm.coordinator.clients

import build.linea.clients.StateManagerV1JsonRpcClient
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapError
import io.vertx.core.Vertx
import net.consensys.linea.BlockInterval
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import net.consensys.linea.jsonrpc.client.JsonRpcRequestRetryer
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class Type2StateManagerJsonRpcClient(
  private val delegate: StateManagerClientV1
) : Type2StateManagerClient {

  data class Config(
    val requestRetry: RequestRetryConfig,
    val zkStateManagerVersion: String
  )

  constructor(
    vertx: Vertx,
    rpcClient: JsonRpcClient,
    config: Config,
    retryConfig: RequestRetryConfig,
    log: Logger = LogManager.getLogger(Type2StateManagerJsonRpcClient::class.java)
  ) : this(
    StateManagerV1JsonRpcClient(
      rpcClient = JsonRpcRequestRetryer(
        vertx,
        rpcClient,
        config = JsonRpcRequestRetryer.Config(
          methodsToRetry = retryableMethods,
          requestRetry = retryConfig
        ),
        log = log
      ),
      config = StateManagerV1JsonRpcClient.Config(
        zkStateManagerVersion = config.zkStateManagerVersion,
        requestRetry = config.requestRetry
      )
    )
  )

  override fun rollupGetZkEVMStateMerkleProof(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<Type2StateManagerErrorType>>> {
    return delegate.rollupGetZkEVMStateMerkleProof(BlockInterval(startBlockNumber.toULong(), endBlockNumber.toULong()))
      .thenApply {
        it.mapError { error -> ErrorResponse(error.type.toLegacyType(), error.message) }
      }
  }

  companion object {
    internal val retryableMethods = setOf("rollup_getZkEVMStateMerkleProofV0")

    fun StateManagerErrorType.toLegacyType(): Type2StateManagerErrorType {
      return when (this) {
        StateManagerErrorType.UNKNOWN -> Type2StateManagerErrorType.UNKNOWN
        StateManagerErrorType.UNSUPPORTED_VERSION -> Type2StateManagerErrorType.UNSUPPORTED_VERSION
        StateManagerErrorType.BLOCK_MISSING_IN_CHAIN -> Type2StateManagerErrorType.BLOCK_MISSING_IN_CHAIN
      }
    }
  }
}
