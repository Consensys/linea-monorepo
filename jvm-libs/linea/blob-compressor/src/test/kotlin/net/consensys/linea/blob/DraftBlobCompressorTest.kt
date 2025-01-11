package net.consensys.linea.blob

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random
import kotlin.concurrent.thread

class DraftBlobCompressorTest {
    private lateinit var compressor: DraftBlobCompressor
    private val dataLimit: UInt = 1024u * 1024u // 1MB

    @BeforeEach
    fun setUp() {
        compressor = DraftBlobCompressor.getInstance(BlobCompressorVersion.V1_0_1, dataLimit)
    }

    @Test
    fun `test successful block compression`() {
        val testData = Random.nextBytes(1000)
        
        // Check if block can be appended
        assertTrue(compressor.canAppendBlock(testData))
        
        // Append block
        val result = compressor.appendBlock(testData)
        assertTrue(result.blockAppended)
        assertTrue(result.compressedSizeAfter > result.compressedSizeBefore)
        
        // Verify compressed data
        val compressedData = compressor.getCompressedData()
        assertNotNull(compressedData)
        assertTrue(compressedData.isNotEmpty())
    }

    @Test
    fun `test compression with data exceeding limit`() {
        val testData = Random.nextBytes(dataLimit.toInt() + 1000)
        
        // Check if block can be appended
        assertFalse(compressor.canAppendBlock(testData))
        
        // Try to append block
        assertThrows<BlobCompressionException> {
            compressor.appendBlock(testData)
        }
    }

    @Test
    fun `test multiple block compression`() {
        val block1 = Random.nextBytes(500)
        val block2 = Random.nextBytes(500)
        
        // Append first block
        val result1 = compressor.appendBlock(block1)
        assertTrue(result1.blockAppended)
        
        // Start new batch
        compressor.startNewBatch()
        
        // Append second block
        val result2 = compressor.appendBlock(block2)
        assertTrue(result2.blockAppended)
        
        // Verify final compressed data
        val compressedData = compressor.getCompressedData()
        assertTrue(compressedData.size > result1.compressedSizeAfter)
    }

    @Test
    fun `test reset functionality`() {
        val testData = Random.nextBytes(1000)
        
        // Append block
        val result = compressor.appendBlock(testData)
        assertTrue(result.blockAppended)
        
        // Reset compressor
        compressor.reset()
        
        // Verify state is cleared
        val compressedData = compressor.getCompressedData()
        assertEquals(0, compressedData.size)
    }

    @Test
    fun `test draft isolation`() {
        val block1 = Random.nextBytes(500)
        val block2 = Random.nextBytes(dataLimit.toInt()) // Block that won't fit
        
        // Append first block successfully
        val result1 = compressor.appendBlock(block1)
        assertTrue(result1.blockAppended)
        val sizeAfterBlock1 = compressor.getCompressedData().size
        
        // Try to append second block (should fail)
        assertFalse(compressor.canAppendBlock(block2))
        
        // Verify original data is preserved
        val finalData = compressor.getCompressedData()
        assertEquals(sizeAfterBlock1, finalData.size)
    }

    @Test
    fun `test concurrent compression using draft pool`() {
        val numThreads = 5
        val blockSize = 500
        val threads = List(numThreads) { threadId ->
            thread {
                repeat(3) {
                    val block = Random.nextBytes(blockSize)
                    assertTrue(compressor.canAppendBlock(block))
                    val result = compressor.appendBlock(block)
                    assertTrue(result.blockAppended)
                }
            }
        }
        
        // Wait for all threads to complete
        threads.forEach { it.join() }
        
        // Verify final state
        val finalData = compressor.getCompressedData()
        assertTrue(finalData.isNotEmpty())
    }

    @Test
    fun `test draft pool reuse`() {
        val blocks = List(10) { Random.nextBytes(500) }
        
        // Process blocks sequentially to test pool reuse
        blocks.forEach { block ->
            assertTrue(compressor.canAppendBlock(block))
            val result = compressor.appendBlock(block)
            assertTrue(result.blockAppended)
            compressor.startNewBatch()
        }
        
        // Verify final state
        val finalData = compressor.getCompressedData()
        assertTrue(finalData.isNotEmpty())
    }
} 
