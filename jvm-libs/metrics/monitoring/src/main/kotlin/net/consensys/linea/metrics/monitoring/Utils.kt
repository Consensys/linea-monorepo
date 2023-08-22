package net.consensys.linea.metrics.monitoring

fun elapsedTimeInMillisSince(startTime: Long): Long = (System.nanoTime() - startTime) / 1_000_000
