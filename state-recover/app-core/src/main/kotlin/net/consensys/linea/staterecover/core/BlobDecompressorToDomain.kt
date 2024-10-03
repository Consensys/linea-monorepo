package net.consensys.linea.staterecover.core

import net.consensys.linea.staterecover.domain.BlockL1RecoveredData

interface BlobDecompressorToDomain {
  fun decompress(blobs: List<ByteArray>): List<BlockL1RecoveredData>
}
