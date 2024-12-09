package linea.rlp

import io.vertx.core.Vertx
import linea.domain.Block
import net.consensys.linea.async.toSafeFuture
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.Callable

class BesuRLPBlockEncoder(
  val vertx: Vertx
) : RLPBlockEncoderAsync {

  override fun encodeAsync(block: Block): SafeFuture<ByteArray> {
    return vertx.executeBlocking(
      Callable {
        RLP.encode(block)
      },
      false
    )
      .toSafeFuture()
  }

  override fun encodeAsync(blocks: List<Block>): SafeFuture<List<ByteArray>> {
    return SafeFuture.collectAll(blocks.map { encodeAsync(it) }.stream())
  }

  override fun decodeAsync(block: ByteArray): SafeFuture<Block> {
    return vertx.executeBlocking(
      Callable {
        RLP.decode(block)
      },
      false
    )
      .toSafeFuture()
  }

  override fun decodeAsync(blocks: List<ByteArray>): SafeFuture<List<Block>> {
    return SafeFuture.collectAll(blocks.map { decodeAsync(it) }.stream())
  }
}
