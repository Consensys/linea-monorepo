package net.consensys.zkevm.coordinator.app.conflationbacktesting

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import kotlin.io.path.Path

class ConflationBacktestingAppTest {

  @Test
  fun `path is updated correctly`() {
    val currentPath = "/shared/v3/prover-execution/requests"
    val updatedPath = ConflationBacktestingApp.getUpdatedPath(
      path = Path(currentPath),
      suffix = "10-20-123456",
    )
    assertEquals("/shared/v3/prover-execution/requests-10-20-123456", updatedPath.toString())
  }
}
