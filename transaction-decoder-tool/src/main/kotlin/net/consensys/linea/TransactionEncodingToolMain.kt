package net.consensys.linea

import build.linea.web3j.domain.toWeb3j
import linea.web3j.toDomain
import net.consensys.linea.BlockParameter.Companion.toBlockParameter
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.utils.Async

class TransactionEncodingToolMain {

  companion object {
    private val log = LogManager.getLogger(TransactionEncodingToolMain::class)

    @JvmStatic
    fun main(args: Array<String>) {
      startApp()
    }

    private fun startApp() {
      val web3j: Web3j = Web3j.build(
        HttpService("https://linea-sepolia.infura.io/v3/"),
        1000,
        Async.defaultExecutorService()
      )
      web3j.ethGetBlockByNumber(924973UL.toBlockParameter().toWeb3j(), true).sendAsync().get().block.toDomain()
    }
  }
}
