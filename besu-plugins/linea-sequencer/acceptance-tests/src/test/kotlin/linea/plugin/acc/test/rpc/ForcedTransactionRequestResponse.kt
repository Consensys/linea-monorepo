/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test.rpc

import com.fasterxml.jackson.core.JsonParser
import com.fasterxml.jackson.core.JsonToken
import com.fasterxml.jackson.databind.DeserializationContext
import com.fasterxml.jackson.databind.JsonDeserializer
import com.fasterxml.jackson.databind.ObjectReader
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests
import org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction
import org.web3j.protocol.core.Request
import java.io.IOException

/**
 * TODO: Replace with a reusable client
 * Parameters for linea_sendForcedRawTransaction RPC call.
 */
data class ForcedTransactionParam(
  val transaction: String,
  val deadline: String,
)

/**
 * Request to send forced transactions.
 */
class SendForcedRawTransactionRequest(
  private val params: List<ForcedTransactionParam>,
) : Transaction<SendForcedRawTransactionResponse> {

  override fun execute(nodeRequests: NodeRequests): SendForcedRawTransactionResponse {
    return try {
      Request(
        "linea_sendForcedRawTransaction",
        listOf(params),
        nodeRequests.web3jService,
        SendForcedRawTransactionResponse::class.java,
      ).send()
    } catch (e: IOException) {
      throw RuntimeException(e)
    }
  }
}

/**
 * Response from linea_sendForcedRawTransaction RPC call.
 * Contains an array of transaction hashes.
 */
class SendForcedRawTransactionResponse : org.web3j.protocol.core.Response<List<String>>()

/**
 * Request to get forced transaction inclusion status.
 */
class GetForcedTransactionInclusionStatusRequest(
  private val txHash: String,
) : Transaction<GetForcedTransactionInclusionStatusResponse> {

  override fun execute(nodeRequests: NodeRequests): GetForcedTransactionInclusionStatusResponse {
    return try {
      Request(
        "linea_getForcedTransactionInclusionStatus",
        listOf(txHash),
        nodeRequests.web3jService,
        GetForcedTransactionInclusionStatusResponse::class.java,
      ).send()
    } catch (e: IOException) {
      throw RuntimeException(e)
    }
  }
}

/**
 * Response from linea_getForcedTransactionInclusionStatus RPC call.
 */
class GetForcedTransactionInclusionStatusResponse :
  org.web3j.protocol.core.Response<GetForcedTransactionInclusionStatusResponse.InclusionStatus>() {

  @Override
  @JsonDeserialize(using = InclusionStatusDeserializer::class)
  override fun setResult(result: InclusionStatus?) {
    super.setResult(result)
  }

  data class InclusionStatus(
    val blockNumber: String,
    val blockTimestamp: String,
    val from: String,
    val inclusionResult: String,
    val transactionHash: String,
  )
}

class InclusionStatusDeserializer :
  JsonDeserializer<GetForcedTransactionInclusionStatusResponse.InclusionStatus>() {

  private val objectReader: ObjectReader = jacksonObjectMapper().reader()

  override fun deserialize(
    jsonParser: JsonParser,
    deserializationContext: DeserializationContext,
  ): GetForcedTransactionInclusionStatusResponse.InclusionStatus? {
    return if (jsonParser.currentToken != JsonToken.VALUE_NULL) {
      objectReader.readValue(
        jsonParser,
        GetForcedTransactionInclusionStatusResponse.InclusionStatus::class.java,
      )
    } else {
      null
    }
  }
}
