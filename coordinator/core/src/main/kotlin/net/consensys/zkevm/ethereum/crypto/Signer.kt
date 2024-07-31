package net.consensys.zkevm.ethereum.crypto

import org.apache.tuweni.bytes.Bytes
import java.math.BigInteger

interface Signer {
  fun sign(bytes: Bytes): Pair<BigInteger, BigInteger>
}
