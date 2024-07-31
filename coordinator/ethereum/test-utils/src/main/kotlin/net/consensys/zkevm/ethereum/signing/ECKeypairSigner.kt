package net.consensys.zkevm.ethereum.signing

import net.consensys.zkevm.ethereum.crypto.Signer
import org.apache.tuweni.bytes.Bytes
import org.web3j.crypto.ECKeyPair
import java.math.BigInteger

class ECKeypairSigner(private val keyPair: ECKeyPair) : Signer {
  override fun sign(bytes: Bytes): Pair<BigInteger, BigInteger> {
    val signature = keyPair.sign(bytes.toArray())
    return signature.r to signature.s
  }
}
