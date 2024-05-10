package net.consensys.zkevm.ethereum.submission

import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class Eip4844SwitchAwareBlobSubmitterTest {
  private lateinit var mockBlobSubmitterAsCallData: BlobSubmitter
  private lateinit var mockBlobSubmitterAsEIP4844: BlobSubmitter
  private lateinit var eip4844SwitchAwareBlobSubmitter: Eip4844SwitchAwareBlobSubmitter

  @BeforeEach
  fun beforeEach() {
    mockBlobSubmitterAsCallData = mock<BlobSubmitter>()
    whenever(mockBlobSubmitterAsCallData.submitBlobCall(any())).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(mockBlobSubmitterAsCallData.submitBlob(any())).thenReturn(SafeFuture.completedFuture(""))
    mockBlobSubmitterAsEIP4844 = mock<BlobSubmitter>()
    whenever(mockBlobSubmitterAsEIP4844.submitBlobCall(any())).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(mockBlobSubmitterAsEIP4844.submitBlob(any())).thenReturn(SafeFuture.completedFuture(""))

    eip4844SwitchAwareBlobSubmitter = Eip4844SwitchAwareBlobSubmitter(
      mockBlobSubmitterAsCallData,
      mockBlobSubmitterAsEIP4844
    )
  }

  @Test
  fun `when eip4844 enabled blob use eip4844 blob submitter`() {
    val mockBlobRecord = mock<BlobRecord>()
    val mockCompressionProof = mock<BlobCompressionProof>()
    whenever(mockCompressionProof.eip4844Enabled).thenReturn(true)
    whenever(mockBlobRecord.blobCompressionProof).thenReturn(mockCompressionProof)

    eip4844SwitchAwareBlobSubmitter.submitBlobCall(mockBlobRecord).get()
    eip4844SwitchAwareBlobSubmitter.submitBlob(mockBlobRecord).get()
    verify(mockBlobSubmitterAsCallData, times(0)).submitBlobCall(any())
    verify(mockBlobSubmitterAsCallData, times(0)).submitBlob(any())
    verify(mockBlobSubmitterAsEIP4844, times(1)).submitBlobCall(mockBlobRecord)
    verify(mockBlobSubmitterAsEIP4844, times(1)).submitBlob(mockBlobRecord)
  }

  @Test
  fun `when eip4844 not enabled blob use call data blob submitter`() {
    val mockBlobRecord = mock<BlobRecord>()
    val mockCompressionProof = mock<BlobCompressionProof>()
    whenever(mockCompressionProof.eip4844Enabled).thenReturn(false)
    whenever(mockBlobRecord.blobCompressionProof).thenReturn(mockCompressionProof)

    eip4844SwitchAwareBlobSubmitter.submitBlobCall(mockBlobRecord).get()
    eip4844SwitchAwareBlobSubmitter.submitBlob(mockBlobRecord).get()
    verify(mockBlobSubmitterAsCallData, times(1)).submitBlobCall(mockBlobRecord)
    verify(mockBlobSubmitterAsCallData, times(1)).submitBlob(mockBlobRecord)
    verify(mockBlobSubmitterAsEIP4844, times(0)).submitBlobCall(any())
    verify(mockBlobSubmitterAsEIP4844, times(0)).submitBlob(any())
  }
}
