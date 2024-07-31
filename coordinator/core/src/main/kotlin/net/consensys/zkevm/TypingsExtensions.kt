package net.consensys.zkevm

import tech.pegasys.teku.infrastructure.unsigned.UInt64

fun UInt64.toULong(): ULong = this.toString().toULong()
