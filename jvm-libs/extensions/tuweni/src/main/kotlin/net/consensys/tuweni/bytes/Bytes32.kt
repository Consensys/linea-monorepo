package net.consensys.tuweni.bytes

import org.apache.tuweni.bytes.Bytes32

fun ByteArray.toBytes32(): Bytes32 = Bytes32.wrap(this)
fun ByteArray.sliceAsBytes32(sliceIndex: Int): Bytes32 = Bytes32.wrap(this, /*offset*/sliceIndex * 32)
