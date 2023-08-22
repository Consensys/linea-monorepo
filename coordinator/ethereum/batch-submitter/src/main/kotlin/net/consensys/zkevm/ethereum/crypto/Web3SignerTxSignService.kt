package net.consensys.zkevm.ethereum.crypto

import org.web3j.crypto.Credentials
import org.web3j.crypto.RawTransaction
import org.web3j.service.TxSignService
import org.web3j.tx.ChainId

class Web3SignerTxSignService(private val credentials: Credentials) : TxSignService {
  override fun sign(rawTransaction: RawTransaction, chainId: Long): ByteArray {
    val signedMessage: ByteArray = if (chainId > ChainId.NONE) {
      Web3SignerFriendlyTransactionEncoder.signMessage(rawTransaction, chainId, credentials)
    } else {
      Web3SignerFriendlyTransactionEncoder.signMessage(rawTransaction, credentials)
    }
    return signedMessage
  }

  override fun getAddress(): String {
    return credentials.address
  }
}
