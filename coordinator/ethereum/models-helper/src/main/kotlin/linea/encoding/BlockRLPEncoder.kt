package linea.encoding

import linea.rlp.RLP
import net.consensys.zkevm.encoding.BlockEncoder

object BlockRLPEncoder : BlockEncoder {
  override fun encode(block: linea.domain.Block): ByteArray = RLP.encode(block)
}
