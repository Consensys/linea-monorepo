package net.consensys.zkevm.encoding

import net.consensys.zkevm.toULong
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPInput
import org.junit.jupiter.api.Test
import tech.pegasys.teku.ethereum.executionclient.schema.randomExecutionPayload

class ExecutionPayloadV1RLPEncoderByBesuImplementationTest {

  @Test
  fun encode() {
    val payload = randomExecutionPayload()
    val rlpEncodedPayload = ExecutionPayloadV1RLPEncoderByBesuImplementation.encode(payload)
    val block = Block.readFrom(BytesValueRLPInput(Bytes.wrap(rlpEncodedPayload), false), MainnetBlockHeaderFunctions())

    assertThat(block.header.number.toULong()).isEqualTo(payload.blockNumber.toULong())
    // we cannot assert oh block hash because Besu will calculate real Hash whereas random payload has random bytes
    // assertThat(block.header.blockHash.toHexString()).isEqualTo(payload.blockHash.toHexString())
    assertThat(block.header.gasLimit.toULong()).isEqualTo(payload.gasLimit.toULong())
    assertThat(block.header.logsBloom.toArray()).isEqualTo(payload.logsBloom.toArray())
    assertThat(block.header.parentHash.toArray()).isEqualTo(payload.parentHash.toArray())
    assertThat(block.header.prevRandao.get().toArray()).isEqualTo(payload.prevRandao.toArray())
    assertThat(block.header.stateRoot.toArray())
      .isEqualTo(payload.stateRoot.toArray())
    assertThat(payload.transactions).isEmpty()

    // FIXME: add remaining fields assertions
  }
}
