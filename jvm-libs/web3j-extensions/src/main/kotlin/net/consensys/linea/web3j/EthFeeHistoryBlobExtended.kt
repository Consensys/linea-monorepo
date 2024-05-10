package net.consensys.linea.web3j

import com.fasterxml.jackson.core.JsonParser
import com.fasterxml.jackson.core.JsonToken
import com.fasterxml.jackson.databind.DeserializationContext
import com.fasterxml.jackson.databind.JsonDeserializer
import com.fasterxml.jackson.databind.ObjectReader
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import net.consensys.linea.FeeHistory
import net.consensys.linea.bigIntFromPrefixedHex
import org.web3j.protocol.ObjectMapperFactory
import org.web3j.protocol.Web3jService
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.Response
import org.web3j.utils.Numeric
import java.math.BigInteger

class EthFeeHistoryBlobExtended : Response<EthFeeHistoryBlobExtended.FeeHistoryBlobExtended>() {
  @JsonDeserialize(using = ResponseDeserialiser::class)
  override fun setResult(result: FeeHistoryBlobExtended) {
    super.setResult(result)
  }

  val feeHistory: FeeHistoryBlobExtended
    get() = if (result != null) super.getResult() else FeeHistoryBlobExtended()

  open class FeeHistoryBlobExtended {
    lateinit var oldestBlock: String
    lateinit var reward: List<List<String>>
    lateinit var baseFeePerGas: List<String>
    lateinit var gasUsedRatio: List<Double>
    var baseFeePerBlobGas: List<String> = emptyList()
    var blobGasUsedRatio: List<Double> = emptyList()
    constructor()

    constructor(
      oldestBlock: String,
      reward: List<List<String>>,
      baseFeePerGas: List<String>,
      gasUsedRatio: List<Double>,
      baseFeePerBlobGas: List<String> = mutableListOf(),
      blobGasUsedRatio: List<Double> = mutableListOf()
    ) {
      this.oldestBlock = oldestBlock
      this.reward = reward
      this.baseFeePerGas = baseFeePerGas
      this.gasUsedRatio = gasUsedRatio
      this.baseFeePerBlobGas = baseFeePerBlobGas
      this.blobGasUsedRatio = blobGasUsedRatio
    }

    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as FeeHistoryBlobExtended

      if (oldestBlock != other.oldestBlock) return false
      if (reward != other.reward) return false
      if (baseFeePerGas != other.baseFeePerGas) return false
      if (gasUsedRatio != other.gasUsedRatio) return false
      if (baseFeePerBlobGas != other.baseFeePerBlobGas) return false
      if (blobGasUsedRatio != other.blobGasUsedRatio) return false

      return true
    }

    override fun hashCode(): Int {
      var result = oldestBlock.hashCode()
      result = 31 * result + reward.hashCode()
      result = 31 * result + baseFeePerGas.hashCode()
      result = 31 * result + gasUsedRatio.hashCode()
      result = 31 * result + baseFeePerBlobGas.hashCode()
      result = 31 * result + blobGasUsedRatio.hashCode()
      return result
    }

    fun toLineaDomain(): FeeHistory {
      return FeeHistory(
        oldestBlock = oldestBlock.bigIntFromPrefixedHex(),
        baseFeePerGas = baseFeePerGas.map { it.bigIntFromPrefixedHex() },
        reward = reward.map { it.map { it.bigIntFromPrefixedHex() } },
        gasUsedRatio = gasUsedRatio.map { it.toBigDecimal() },
        baseFeePerBlobGas = baseFeePerBlobGas.map { it.bigIntFromPrefixedHex() },
        blobGasUsedRatio = blobGasUsedRatio.map { it.toBigDecimal() }
      )
    }
  }

  class ResponseDeserialiser : JsonDeserializer<FeeHistoryBlobExtended>() {
    private val objectReader: ObjectReader = ObjectMapperFactory.getObjectReader()

    override fun deserialize(
      jsonParser: JsonParser,
      deserializationContext: DeserializationContext
    ): FeeHistoryBlobExtended? {
      return if (jsonParser.currentToken != JsonToken.VALUE_NULL) {
        objectReader.readValue(jsonParser, FeeHistoryBlobExtended::class.java)
      } else {
        null
      }
    }
  }
}

class Web3jBlobExtended(private val web3jService: Web3jService) {
  fun ethFeeHistoryWithBlob(
    blockCount: Int,
    newestBlock: DefaultBlockParameter,
    rewardPercentiles: List<Double>
  ): Request<*, EthFeeHistoryBlobExtended> {
    return Request(
      "eth_feeHistory",
      listOf(
        Numeric.encodeQuantity(BigInteger.valueOf(blockCount.toLong())),
        newestBlock.value,
        rewardPercentiles
      ),
      this.web3jService,
      EthFeeHistoryBlobExtended::class.java
    )
  }
}
