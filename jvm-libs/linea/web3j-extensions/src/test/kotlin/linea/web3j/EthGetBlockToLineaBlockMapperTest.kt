package linea.web3j

import io.vertx.core.json.JsonObject
import linea.domain.Transaction
import linea.domain.TransactionType
import linea.domain.toBesu
import net.consensys.decodeHex
import net.consensys.toBigInteger
import net.consensys.toBigIntegerFromHex
import net.consensys.toLongFromHex
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Wei
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.methods.response.EthBlock
import kotlin.jvm.optionals.getOrNull

class EthGetBlockToLineaBlockMapperTest {

  @Test
  fun `should map frontier transactions`() {
    val transactionJson = JsonObject(
      // example of eth_getBlockByNumber.transactions[0] response
      // This seems to be easier to read given than build Java object directly
      """
      {
        "blockHash": "0x004257e560a5f82595dddb73f752b904efef4b73cb3ece1469f5e5091e3c9665",
        "blockNumber": "0xe1d30",
        "chainId": "0xe705",
        "from": "0x228466f2c715cbec05deabfac040ce3619d7cf0b",
        "gas": "0x5208",
        "gasPrice": "0xee2d984",
        "hash": "0x5d3b5e1ae3e4ea5612e6907cb09c4e0e5482171b4c2af794e17b77314547bb79",
        "input": "0x",
        "nonce": "0x97411",
        "r": "0xdf28597129341d5d345c9043c7d0b0a22be82cac13988cfc1d8cbdaf3ab3f35b",
        "s": "0x3189b2ff80d8f728d6fb7503b46734ee77a60a42db01d0b09db10bdc9d5caa44",
        "to": "0x228466f2c715cbec05deabfac040ce3619d7cf0b",
        "transactionIndex": "0x0",
        "type": "0x0",
        "v": "0x1ce2e",
        "value": "0x186a0"
      }
      """.trimIndent()
    )
    val web3jTx = EthBlock.TransactionObject(
      /*hash*/ transactionJson.getString("hash"),
      /*nonce*/ transactionJson.getString("nonce"),
      /*blockHash*/ transactionJson.getString("blockHash"),
      /*blockNumber*/ transactionJson.getString("blockNumber"),
      /*chainId*/ transactionJson.getString("chainId"),
      /*transactionIndex*/ transactionJson.getString("transactionIndex"),
      /*from*/ transactionJson.getString("from"),
      /*to*/ transactionJson.getString("to"),
      /*value*/ transactionJson.getString("value"),
      /*gasPrice*/ transactionJson.getString("gasPrice"),
      /*gas*/ transactionJson.getString("gas"),
      /*input*/ transactionJson.getString("input"),
      /*creates*/ null,
      /*publicKey*/null,
      /*raw*/null,
      /*r*/ transactionJson.getString("r"),
      /*s*/ transactionJson.getString("s"),
      /*v*/ transactionJson.getString("v").toLongFromHex(),
      /*yParity*/null,
      /*type*/ transactionJson.getString("type"),
      /*maxFeePerGas*/null,
      /*maxPriorityFeePerGas*/ null,
      /*> accessList*/emptyList()
    )

    val domainTx = web3jTx.toDomain()
    assertThat(domainTx).isEqualTo(
      Transaction(
        nonce = 0x97411UL,
        gasPrice = 0xee2d984UL,
        gasLimit = 0x5208UL,
        to = "0x228466f2c715cbec05deabfac040ce3619d7cf0b".decodeHex(),
        value = 0x186a0UL.toBigInteger(),
        input = "0x".decodeHex(),
        r = "0xdf28597129341d5d345c9043c7d0b0a22be82cac13988cfc1d8cbdaf3ab3f35b".toBigIntegerFromHex(),
        s = "0x3189b2ff80d8f728d6fb7503b46734ee77a60a42db01d0b09db10bdc9d5caa44".toBigIntegerFromHex(),
        v = 118318UL,
        yParity = null,
        type = TransactionType.FRONTIER,
        chainId = 0xe705UL,
        maxFeePerGas = null,
        maxPriorityFeePerGas = null
      )
    )
    domainTx.toBesu().also { besuTx ->
      assertThat(besuTx.nonce).isEqualTo(0x97411L)
      assertThat(besuTx.gasPrice.getOrNull()).isEqualTo(Wei.of(0xee2d984L))
      assertThat(besuTx.gasLimit).isEqualTo(0x5208L)
      assertThat(besuTx.to.getOrNull()).isEqualTo(Address.fromHexString("0x228466f2c715cbec05deabfac040ce3619d7cf0b"))
      assertThat(besuTx.value).isEqualTo(Wei.of(0x186a0L))
      assertThat(besuTx.payload).isEqualTo(Bytes.EMPTY)
      assertThat(besuTx.signature.r).isEqualTo(
        "0xdf28597129341d5d345c9043c7d0b0a22be82cac13988cfc1d8cbdaf3ab3f35b".toBigIntegerFromHex()
      )
      assertThat(besuTx.signature.s).isEqualTo(
        "0x3189b2ff80d8f728d6fb7503b46734ee77a60a42db01d0b09db10bdc9d5caa44".toBigIntegerFromHex()
      )
      assertThat(besuTx.signature.recId).isEqualTo(1)
      assertThat(besuTx.type).isEqualTo(org.hyperledger.besu.datatypes.TransactionType.FRONTIER)
      assertThat(besuTx.chainId.getOrNull()).isEqualTo(0xe705L)
      assertThat(besuTx.maxFeePerGas).isEmpty()
      assertThat(besuTx.maxPriorityFeePerGas).isEmpty()
    }
  }
}
