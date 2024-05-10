package net.consensys.zkevm.load

import net.consensys.zkevm.load.model.EthConnection
import net.consensys.zkevm.load.model.TransactionDetail
import net.consensys.zkevm.load.model.Wallet
import org.slf4j.LoggerFactory
import org.web3j.crypto.RawTransaction
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.utils.Numeric
import java.io.IOException
import java.math.BigInteger
import java.security.SecureRandom
import java.util.function.Consumer
import kotlin.collections.ArrayList

class WalletsFunding(
  private val ethConnection: EthConnection,
  private val sourceOfFundsWallet: Wallet
) {

  val numberGenerator = SecureRandom()

  @Throws(IOException::class)
  fun generateTxWithRandomPayload(
    wallets: Map<Int, Wallet>,
    payloadSize: Int,
    chainId: Int,
    costPerCall: BigInteger,
    nbTransfers: Int
  ): Map<Wallet, List<TransactionDetail>> {
    val randomBytes = ByteArray(payloadSize * 2)
    numberGenerator.nextBytes(randomBytes)
    val largePayLoad = String(randomBytes)
    return generateTxsWithPayload(
      wallets,
      largePayLoad.substring(0, payloadSize.coerceAtMost(largePayLoad.length)),
      chainId,
      costPerCall,
      nbTransfers
    )
  }

  fun generateTxsWithPayload(
    wallets: Map<Int, Wallet>,
    payLoad: String,
    chainId: Int,
    costPerCall: BigInteger,
    nbTransfers: Int
  ): MutableMap<Wallet, List<TransactionDetail>> {
    val gasPrice = ethConnection.ethGasPrice()
    val result: MutableMap<Wallet, List<TransactionDetail>> = HashMap()

    for (sourceWallet in wallets.values) {
      val txs = ArrayList<TransactionDetail>()
      for (i in 0 until nbTransfers) {
        val rawTransaction = RawTransaction.createTransaction(
          sourceWallet.theoreticalNonceValue,
          gasPrice,
          costPerCall,
          Numeric.prependHexPrefix(sourceWallet.address),
          BigInteger.ZERO,
          payLoad
        )
        txs.add(
          TransactionDetail(
            sourceWallet.id,
            sourceWallet.theoreticalNonceValue,
            ethConnection.ethSendRawTransaction(rawTransaction, sourceWallet, chainId),
            costPerCall
          )
        )
        sourceWallet.incrementTheoreticalNonce()
      }
      result[sourceWallet] = txs
    }

    return result
  }

  @Throws(IOException::class)
  fun generateUnderPricedTxs(
    wallets: Map<Int, Wallet>,
    chainId: Int
  ): Map<Wallet, List<TransactionDetail>> {
    val gasUnderPriced = ethConnection.ethGasPrice().multiply(BigInteger.valueOf(80))
      .divide(BigInteger.valueOf(100))
    val result: MutableMap<Wallet, List<TransactionDetail>> = HashMap()
    val costPerCall = EthConnection.SIMPLE_TX_PRICE
    logger.debug("[UNPROFITABLE] estimated cost per call: {}", costPerCall)
    for (sourceWallet in wallets.values) {
      val rawTransaction = RawTransaction.createTransaction(
        sourceWallet.theoreticalNonceValue,
        gasUnderPriced,
        costPerCall,
        Numeric.prependHexPrefix(sourceWallet.address),
        BigInteger.ZERO,
        null
      )
      result[sourceWallet] = java.util.List.of(
        TransactionDetail(
          sourceWallet.id,
          sourceWallet.theoreticalNonceValue,
          ethConnection.ethSendRawTransaction(rawTransaction, sourceWallet, chainId),
          EthConnection.SIMPLE_TX_PRICE
        )
      )
      sourceWallet.incrementTheoreticalNonce()
    }
    return result
  }

  // add some eth to the wallets. Needs to be run once when the wallets are created, or when the balance decreased too much (because of paid gas)
  @Throws(IOException::class, InterruptedException::class)
  fun initializeWallets(
    wallets: Map<Int, Wallet>,
    nbTransactions: Int,
    sourceWallet: Wallet,
    chainId: Int,
    costPerCall: BigInteger
  ): Map<Wallet, List<TransactionDetail>> {
    val balance = ethConnection.getBalance(sourceWallet)
    logger.info("[FUNDING] source of funds balance is {}.", balance)
    var nonce = ethConnection.getNonce(sourceOfFundsWallet.encodedAddress())
    logger.info("[FUNDING] initial nonce is {}.", nonce)
    val gasPrice = ethConnection.ethGasPrice()
    logger.debug("[FUNDING] estimated cost per call: {}", costPerCall)
    val transferredAmount =
      gasPrice.multiply(BigInteger.valueOf(nbTransactions.toLong())).multiply(costPerCall)
    logger.debug(
      "[FUNDING] gas price is {}, transferred amount is {}.",
      gasPrice,
      transferredAmount
    )
    val fundingTransfers: ArrayList<TransactionDetail> = ArrayList()
    for ((_, value) in wallets) {
      val rawTx = createFundTransferTransaction(
        sourceWallet,
        value.address,
        nonce,
        gasPrice,
        EthConnection.SIMPLE_TX_PRICE,
        transferredAmount,
        chainId
      )
      fundingTransfers.add(
        TransactionDetail(
          sourceOfFundsWallet.id,
          nonce,
          rawTx,
          EthConnection.SIMPLE_TX_PRICE
        )
      )
      val res = rawTx.send()
      logger.debug(
        "[FUNDING] Transfer fund transaction sent for nonce:{}, hash:{}, {}.",
        nonce,
        res.transactionHash,
        if (res.error != null) " error:" + res.error.message else "no error"
      )
      nonce = nonce.add(BigInteger.ONE)
    }
    logger.info("[FUNDING] Waiting for fund transfer.")
    while (ethConnection.ethGetTransactionCount(
        sourceOfFundsWallet.encodedAddress(),
        DefaultBlockParameterName.LATEST
      ) < nonce
    ) {
      Thread.sleep(10)
    }
    logger.info("[FUNDING] completed.")

    return mapOf(sourceOfFundsWallet to fundingTransfers)
  }

  private fun createFundTransferTransaction(
    wallet: Wallet,
    toAddress: String,
    nonce: BigInteger,
    gasPrice: BigInteger,
    gasLimit: BigInteger,
    initialAmount: BigInteger,
    chainId: Int
  ): Request<*, EthSendTransaction> {
    val rawTransaction = RawTransaction.createEtherTransaction(
      nonce,
      gasPrice,
      gasLimit,
      Numeric.prependHexPrefix(toAddress),
      initialAmount
    )
    return ethConnection.ethSendRawTransaction(rawTransaction, wallet, chainId)
  }

  @Throws(Exception::class)
  fun generateTransactions(
    wallets: Map<Int, Wallet>,
    valueToTransfer: BigInteger,
    nbTransactions: Int,
    chainId: Int
  ): Map<Wallet, MutableList<TransactionDetail>> {
    // check wallet balance, it helps to ensure wallets exist.
    wallets.values.forEach(Consumer { a: Wallet? -> ethConnection.getBalance(a!!) })
    val initialNoncePerWallet = wallets.entries.associate {
      it.value to ArrayList<TransactionDetail>()
    }
    return getTransactionDetailList(
      wallets,
      valueToTransfer,
      nbTransactions,
      initialNoncePerWallet,
      chainId
    )
  }

  @Throws(InterruptedException::class)
  fun waitForTransactions(targetNonce: Map<Wallet, BigInteger>, timeoutDelay: Long = 600L) {
    ethConnection.waitForTransactions(targetNonce, timeoutDelay)
  }

  @Throws(IOException::class)
  private fun getTransactionDetailList(
    wallets: Map<Int, Wallet>,
    value: BigInteger,
    nbTransactions: Int,
    initialNoncePerWallet: Map<Wallet, List<TransactionDetail>>,
    chainId: Int
  ):
    Map<Wallet, MutableList<TransactionDetail>> {
    val transactions: MutableMap<Wallet, MutableList<TransactionDetail>> = HashMap()
    val gasPrice = ethConnection.ethGasPrice()
    for ((wallet) in initialNoncePerWallet) {
      transactions[wallet] = ArrayList()
      logger.info("[TRANSFER] Preparing transactions for wallet " + wallet.address + " id: " + wallet.id)
      val walletId = wallet.id
      for (i in 0 until nbTransactions) {
        val nonce = wallet.theoreticalNonce.get()
        val toAddress = getDestinationWallet(wallets, wallet, walletId, i)
        val rawTransaction = createFundTransferTransaction(
          wallet,
          toAddress,
          nonce,
          gasPrice,
          EthConnection.SIMPLE_TX_PRICE,
          value,
          chainId
        )
        transactions[wallet]!!.add(
          TransactionDetail(
            walletId,
            nonce,
            rawTransaction,
            EthConnection.SIMPLE_TX_PRICE
          )
        )
        wallet.incrementTheoreticalNonce()
      }
    }
    return transactions
  }

  private fun getDestinationWallet(
    wallets: Map<Int, Wallet>,
    wallet: Wallet,
    walletId: Int,
    i: Int
  ): String {
    val walletDestinationId = (walletId + i + 1) % wallets.size
    logger.debug(
      "[TRANSFER] preparing transactions from wallet {} to wallet {}",
      wallet,
      walletDestinationId
    )
    return wallets[walletDestinationId]!!.address
  }

  companion object {
    val logger = LoggerFactory.getLogger(WalletsFunding::class.java)
  }
}
