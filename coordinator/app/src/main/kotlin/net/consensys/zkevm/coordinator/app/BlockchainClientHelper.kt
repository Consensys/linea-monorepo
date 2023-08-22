package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import io.vertx.core.http.HttpVersion
import io.vertx.ext.web.client.WebClientOptions
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.linea.httprest.client.VertxHttpRestClient
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesCalculator
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesFetcher
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.WMAGasProvider
import net.consensys.zkevm.ethereum.crypto.ECKeypairSignerAdapter
import net.consensys.zkevm.ethereum.crypto.Web3SignerRestClient
import net.consensys.zkevm.ethereum.crypto.Web3SignerTxSignService
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.service.TxSignServiceImpl
import org.web3j.utils.Numeric
import java.net.URI

fun createTransactionManager(
  vertx: Vertx,
  signerConfig: SignerConfig,
  client: Web3j
): AsyncFriendlyTransactionManager {
  val transactionSignService = when (signerConfig.type) {
    SignerConfig.Type.Web3j -> {
      TxSignServiceImpl(Credentials.create(signerConfig.web3j!!.privateKey.value))
    }

    SignerConfig.Type.Web3Signer -> {
      val web3SignerConfig = signerConfig.web3signer!!
      val endpoint = URI(web3SignerConfig.endpoint)
      val webClientOptions: WebClientOptions =
        WebClientOptions()
          .setKeepAlive(web3SignerConfig.keepAlive)
          .setProtocolVersion(HttpVersion.HTTP_1_1)
          .setMaxPoolSize(web3SignerConfig.maxPoolSize.toInt())
          .setDefaultHost(endpoint.host)
          .setDefaultPort(endpoint.port)
      val httpRestClient = VertxHttpRestClient(webClientOptions, vertx)
      val signer = Web3SignerRestClient(httpRestClient, signerConfig.web3signer.publicKey)
      val signerAdapter = ECKeypairSignerAdapter(signer, Numeric.toBigInt(signerConfig.web3signer.publicKey))
      val web3SignerCredentials = Credentials.create(signerAdapter)
      Web3SignerTxSignService(web3SignerCredentials)
    }
  }

  return AsyncFriendlyTransactionManager(client, transactionSignService, client.ethChainId().send().id)
}

fun instantiateZkEvmContractClient(
  l1Config: L1Config,
  transactionManager: AsyncFriendlyTransactionManager,
  gasFetcher: FeesFetcher,
  wmaFeesCalculator: FeesCalculator,
  client: Web3j
): ZkEvmV2AsyncFriendly {
  return ZkEvmV2AsyncFriendly.load(
    l1Config.zkEvmContractAddress,
    client,
    transactionManager,
    WMAGasProvider(
      client.ethChainId().send().chainId.toLong(),
      gasFetcher,
      wmaFeesCalculator,
      WMAGasProvider.Config(
        l1Config.gasLimit,
        l1Config.maxFeePerGas
      )
    )
  )
}
