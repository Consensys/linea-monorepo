package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.getError
import net.consensys.linea.BlockNumberAndHash
import org.apache.tuweni.bytes.Bytes32
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
      BlockNumberAndHash(15u, Bytes32.random()),
      BlockNumberAndHash(14u, Bytes32.random()),
      // gap on 13
      BlockNumberAndHash(12u, Bytes32.random()),
      BlockNumberAndHash(11u, Bytes32.random()),
      BlockNumberAndHash(10u, Bytes32.random())
    )
    assertBlocksList(blocks).let { result ->
      assertThat(result).isInstanceOf(Err::class.java)
      assertThat(result.getError()).isInstanceOf(IllegalArgumentException::class.java)
      assertThat(result.getError()!!.message).isEqualTo("Conflated blocks list has non consecutive blocks!")
    }
  }

  @Test fun `assertBlocksList returns sorted list when all blocks are consecutive`() {
    val blocks = listOf(
      BlockNumberAndHash(13u, Bytes32.random()),
      BlockNumberAndHash(12u, Bytes32.random()),
      BlockNumberAndHash(11u, Bytes32.random()),
      BlockNumberAndHash(10u, Bytes32.random())
    )

    assertThat(assertBlocksList(blocks)).isEqualTo(Ok(blocks.sortedBy { it.number }))
    assertThat(assertBlocksList(listOf(blocks[0]))).isEqualTo(Ok(listOf(blocks[0])))
  }
}
