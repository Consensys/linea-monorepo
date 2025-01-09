package linea.staterecover.plugin

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.BesuPlugin
import org.hyperledger.besu.plugin.ServiceManager

open class JacksonTestPlugin : BesuPlugin {
  private val log: Logger = LogManager.getLogger(JacksonTestPlugin::class.java)

  override fun register(serviceManager: ServiceManager) {
    log.info("JacksonTestPlugin Registered")
  }

  override fun start() {
    log.info("JacksonTestPlugin starting")
    println("jackson loaded" + JacksonHelper.someValue)
  }

  override fun afterExternalServicePostMainLoop() {
  }

  override fun stop() {
    // no-op
  }
}
