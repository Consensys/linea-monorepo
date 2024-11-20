package build.linea.staterecover.clients.el

import build.linea.s11n.jackson.ethApiObjectMapper
import build.linea.staterecover.BlockL1RecoveredData
import build.linea.staterecover.clients.ExecutionLayerClient
import com.fasterxml.jackson.databind.JsonNode
import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.fromHexString
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import net.consensys.linea.jsonrpc.client.JsonRpcClientFactory
import net.consensys.linea.jsonrpc.client.JsonRpcV2Client
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.net.URI

class ExecutionLayerJsonRpcClient internal constructor(
  private val rpcClient: JsonRpcV2Client
) : ExecutionLayerClient {
  override fun getBlockNumberAndHash(blockParameter: BlockParameter): SafeFuture<BlockNumberAndHash> {
    return rpcClient
      .makeRequest(
        method = "eth_getBlockByNumber",
        params = listOf(blockParameter.toString(), false)
      ) { result: Any? ->
        @Suppress("UNCHECKED_CAST")
        result as JsonNode
        BlockNumberAndHash(
          number = ULong.fromHexString(result.get("number").asText()),
          hash = result.get("hash").asText().decodeHex()
        )
      }
  }

  override fun lineaEngineImportBlocksFromBlob(blocks: List<BlockL1RecoveredData>): SafeFuture<Unit> {
    return rpcClient
      .makeRequest(
        method = "linea_engine_importBlocksFromBlob",
        params = blocks,
        resultMapper = { Unit }
      )
  }

  override fun lineaEngineForkChoiceUpdated(
    headBlockHash: ByteArray,
    finalizedBlockHash: ByteArray
  ): SafeFuture<Unit> {
    return rpcClient
      .makeRequest(
        method = "linea_engine_importForkChoiceUpdated",
        params = listOf(headBlockHash, finalizedBlockHash).map { it.encodeHex() },
        resultMapper = { Unit }
      )
  }

  companion object {
    fun create(
      rpcClientFactory: JsonRpcClientFactory,
      endpoint: URI,
      requestRetryConfig: RequestRetryConfig
    ): ExecutionLayerClient {
      return ExecutionLayerJsonRpcClient(
        rpcClient = rpcClientFactory.createJsonRpcV2Client(
          endpoints = listOf(endpoint),
          retryConfig = requestRetryConfig,
          requestObjectMapper = ethApiObjectMapper
        )
      )
    }
  }
}
