package net.consensys.tuweni.bytes

import net.consensys.toULong
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256

fun ByteArray.toBytes32(): Bytes32 = Bytes32.wrap(this)
fun ByteArray.sliceAsBytes32(sliceIndex: Int): Bytes32 = Bytes32.wrap(this, /*offset*/sliceIndex * 32)
fun Bytes32.toULong(): ULong = UInt256.fromBytes(this).toBigInteger().toULong()
