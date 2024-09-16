package net.consensys.linea.transactionexclusion.dto

import com.fasterxml.jackson.databind.ObjectMapper
import net.consensys.linea.transactionexclusion.ModuleOverflow
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.linea.transactionexclusion.app.api.ArgumentParser

data class ModuleOverflowJsonDto(
  val count: Long,
  val limit: Long,
  val module: String
) {
  // Jackson ObjectMapper requires a default constructor
  constructor() : this(0L, 0L, "")

  companion object {
    fun parseListFromJsonString(jsonString: String): List<ModuleOverflowJsonDto> {
      return ObjectMapper().readValue(
        jsonString,
        Array<ModuleOverflowJsonDto>::class.java
      ).toList()
    }

    fun parseToJsonString(target: Any): String {
      if (target is String) {
        return target
      }
      return ObjectMapper().writeValueAsString(target)
    }
  }

  fun toDomainObject(): ModuleOverflow {
    return ModuleOverflow(
      count = count,
      limit = limit,
      module = module
    )
  }
}

data class RejectedTransactionJsonDto(
  val txRejectionStage: String,
  val timestamp: String,
  val blockNumber: String?,
  val transactionRLP: String,
  val reasonMessage: String,
  val overflows: Any
) {
  // Jackson ObjectMapper requires a default constructor
  constructor() : this("", "", null, "", "", Any())

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as RejectedTransactionJsonDto

    if (txRejectionStage != other.txRejectionStage) return false
    if (timestamp != other.timestamp) return false
    if (blockNumber != other.blockNumber) return false
    if (transactionRLP != other.transactionRLP) return false
    if (reasonMessage != other.reasonMessage) return false
    return overflows == other.overflows
  }

  override fun hashCode(): Int {
    var result = txRejectionStage.hashCode()
    result = 31 * result + timestamp.hashCode()
    result = 31 * result + (blockNumber?.hashCode() ?: 0)
    result = 31 * result + transactionRLP.hashCode()
    result = 31 * result + reasonMessage.hashCode()
    result = 31 * result + overflows.hashCode()
    return result
  }

//  companion object {
//    fun parseFromJsonString(jsonString: String): RejectedTransactionJsonDto {
//      return ObjectMapper().readValue(
//        jsonString,
//        RejectedTransactionJsonDto::class.java
//      )
//    }
//
//    fun parseToJsonString(target: Any): String {
//      if (target is String) {
//        return target
//      }
//      return ObjectMapper().writeValueAsString(target)
//    }
//  }

  fun toDomainObject(): RejectedTransaction {
    return RejectedTransaction(
      txRejectionStage = ArgumentParser.getTxRejectionStage(txRejectionStage),
      timestamp = ArgumentParser.getTimestampFromISO8601(timestamp),
      blockNumber = ArgumentParser.getBlockNumber(blockNumber),
      transactionRLP = ArgumentParser.getTransactionRLPInRawBytes(transactionRLP),
      reasonMessage = ArgumentParser.getReasonMessage(reasonMessage),
      overflows = ArgumentParser.getOverflows(overflows)
    ).also {
      it.transactionInfo = ArgumentParser.getTransactionInfoFromRLP(it.transactionRLP)
    }
  }
}
