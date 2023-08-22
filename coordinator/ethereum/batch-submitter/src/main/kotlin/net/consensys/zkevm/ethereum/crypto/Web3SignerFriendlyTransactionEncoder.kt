package net.consensys.zkevm.ethereum.crypto

import org.web3j.crypto.Credentials
import org.web3j.crypto.ECKeyPair
import org.web3j.crypto.Hash
import org.web3j.crypto.RawTransaction
import org.web3j.crypto.Sign
import org.web3j.crypto.TransactionEncoder

object Web3SignerFriendlyTransactionEncoder {
  fun signMessage(
    rawTransaction: RawTransaction,
    chainId: Long,
    credentials: Credentials
  ): ByteArray {
    // Eip1559: Tx has ChainId inside
    if (rawTransaction.type.isEip1559) {
      return signMessage(rawTransaction, credentials)
    }
    val encodedTransaction = TransactionEncoder.encode(rawTransaction, chainId)
    val signatureData = signMessage(encodedTransaction, credentials.ecKeyPair)
    val eip155SignatureData = TransactionEncoder.createEip155SignatureData(signatureData, chainId)
    return TransactionEncoder.encode(rawTransaction, eip155SignatureData)
  }

  fun signMessage(rawTransaction: RawTransaction, credentials: Credentials): ByteArray {
    val encodedTransaction = TransactionEncoder.encode(rawTransaction)
    val signatureData = signMessage(encodedTransaction, credentials.ecKeyPair)
    return TransactionEncoder.encode(rawTransaction, signatureData)
  }

  private fun signMessage(message: ByteArray, keyPair: ECKeyPair): Sign.SignatureData {
    val publicKey = keyPair.publicKey
    val messageHash: ByteArray = Hash.sha3(message)
    val sig = keyPair.sign(message)
    return Sign.createSignatureData(sig, publicKey, messageHash)
  }
}
