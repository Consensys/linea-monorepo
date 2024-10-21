package net.consensys.linea.web3j

import net.consensys.linea.Constants

fun padBlobForEip4844Submission(blob: ByteArray): ByteArray {
  return ByteArray(Constants.Eip4844BlobSize).apply { blob.copyInto(this) }
}
