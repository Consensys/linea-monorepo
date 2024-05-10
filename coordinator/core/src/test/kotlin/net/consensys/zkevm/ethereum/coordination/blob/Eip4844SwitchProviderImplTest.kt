package net.consensys.zkevm.ethereum.coordination.blob

import net.consensys.zkevm.domain.Blob
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock

class Eip4844SwitchProviderImplTest {
  @Test
  fun `when switch block number is -1 eip4844 is not enabled`() {
    val blob = mock<Blob>()
    assertThat(Eip4844SwitchProviderImpl(-1).isEip4844Enabled(blob)).isFalse()
  }

  @Test
  fun `when switch block number less than or equal blob start eip4844 is enabled`() {
    val eip4844SwitchProvider = Eip4844SwitchProviderImpl(11)
    (11uL..20uL).forEach {
      assertThat(eip4844SwitchProvider.isEip4844Enabled(it)).isTrue()
    }
  }

  @Test
  fun `when switch block number more than blob start eip4844 is not enabled`() {
    val eip4844SwitchProvider = Eip4844SwitchProviderImpl(11)
    (0uL..10uL).forEach {
      assertThat(eip4844SwitchProvider.isEip4844Enabled(it)).isFalse()
    }
  }
}
