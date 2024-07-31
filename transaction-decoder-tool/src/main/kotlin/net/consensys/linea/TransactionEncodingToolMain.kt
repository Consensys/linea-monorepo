package net.consensys.linea

import org.apache.logging.log4j.LogManager

class TransactionEncodingToolMain {

  companion object {
    private val log = LogManager.getLogger(TransactionEncodingToolMain::class)

    @JvmStatic
    fun main(args: Array<String>) {
      startApp()
    }

    private fun startApp() {
      val app = BlockReader()
      app.getBlockPayload(924973)
    }
  }
}
