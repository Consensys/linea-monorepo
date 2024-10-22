package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.getError
import net.consensys.ByteArrayE
import net.consensys.linea.BlockNumberAndHash
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class TracesConflationCoordinatorImplTest {

  @Test fun `assertBlocksList return error when empty`() {
    assertBlocksList(emptyList()).let { result ->
      assertThat(result).isInstanceOf(Err::class.java)
      assertThat(result.getError()).isInstanceOf(IllegalArgumentException::class.java)
      assertThat(result.getError()!!.message).isEqualTo("Empty list of blocs")
    }
  }

  @Test fun `assertBlocksList return error when there is gap in block numbers`() {
    val blocks = listOf(
      BlockNumberAndHash(15u, ByteArrayE.random32()),
      BlockNumberAndHash(14u, ByteArrayE.random32()),
      // gap on 13
      BlockNumberAndHash(12u, ByteArrayE.random32()),
      BlockNumberAndHash(11u, ByteArrayE.random32()),
      BlockNumberAndHash(10u, ByteArrayE.random32())
    )
    assertBlocksList(blocks).let { result ->
      assertThat(result).isInstanceOf(Err::class.java)
      assertThat(result.getError()).isInstanceOf(IllegalArgumentException::class.java)
      assertThat(result.getError()!!.message).isEqualTo("Conflated blocks list has non consecutive blocks!")
    }
  }

  @Test fun `assertBlocksList returns sorted list when all blocks are consecutive`() {
    val blocks = listOf(
      BlockNumberAndHash(13u, ByteArrayE.random32()),
      BlockNumberAndHash(12u, ByteArrayE.random32()),
      BlockNumberAndHash(11u, ByteArrayE.random32()),
      BlockNumberAndHash(10u, ByteArrayE.random32())
    )

    assertThat(assertBlocksList(blocks)).isEqualTo(Ok(blocks.sortedBy { it.number }))
    assertThat(assertBlocksList(listOf(blocks[0]))).isEqualTo(Ok(listOf(blocks[0])))
  }
}
