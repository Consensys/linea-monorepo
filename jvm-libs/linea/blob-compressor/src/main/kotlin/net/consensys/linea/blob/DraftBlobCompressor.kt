package net.consensys.linea.blob

import org.apache.logging.log4j.LogManager
import java.util.concurrent.ConcurrentLinkedQueue

/**
 * Improved implementation of BlobCompressor that uses a draft-based approach
 * instead of optimistic compression with revert.
 */
class DraftBlobCompressor private constructor(
    private val goNativeBlobCompressor: GoNativeBlobCompressor,
    private val dataLimit: UInt,
    private val compressorVersion: BlobCompressorVersion
) : BlobCompressor {

    companion object {
        private val log = LogManager.getLogger(DraftBlobCompressor::class.java)
        private const val DRAFT_POOL_SIZE = 3 // Configurable pool size

        @Volatile
        private var instance: DraftBlobCompressor? = null

        fun getInstance(
            compressorVersion: BlobCompressorVersion = BlobCompressorVersion.V1_0_1,
            dataLimit: UInt
        ): DraftBlobCompressor {
            if (instance == null) {
                synchronized(this) {
                    if (instance == null) {
                        val mainCompressor = GoNativeBlobCompressorFactory.getInstance(compressorVersion)
                        
                        // Initialize main compressor
                        val dictionaryPath = GoNativeBlobCompressorFactory.dictionaryPath.toString()
                        if (!mainCompressor.Init(dataLimit.toInt(), dictionaryPath)) {
                            throw InstantiationException(mainCompressor.Error())
                        }
                        
                        instance = DraftBlobCompressor(mainCompressor, dataLimit, compressorVersion)
                    } else {
                        throw IllegalStateException("Compressor singleton instance already created")
                    }
                }
            } else {
                throw IllegalStateException("Compressor singleton instance already created")
            }
            return instance!!
        }
    }

    // Pool of draft compressors
    private val draftPool = ConcurrentLinkedQueue<GoNativeBlobCompressor>()

    init {
        // Initialize pool with draft compressors
        val dictionaryPath = GoNativeBlobCompressorFactory.dictionaryPath.toString()
        repeat(DRAFT_POOL_SIZE) {
            val draft = GoNativeBlobCompressorFactory.getInstance(compressorVersion)
            if (!draft.Init(dataLimit.toInt(), dictionaryPath)) {
                throw InstantiationException(draft.Error())
            }
            draftPool.offer(draft)
        }
    }

    private fun getDraftCompressor(): GoNativeBlobCompressor {
        return draftPool.poll() ?: run {
            // Create new draft compressor if pool is empty
            val draft = GoNativeBlobCompressorFactory.getInstance(compressorVersion)
            val dictionaryPath = GoNativeBlobCompressorFactory.dictionaryPath.toString()
            if (!draft.Init(dataLimit.toInt(), dictionaryPath)) {
                throw InstantiationException(draft.Error())
            }
            draft
        }
    }

    private fun returnDraftCompressor(draft: GoNativeBlobCompressor) {
        draft.Reset()
        draftPool.offer(draft)
    }

    override fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean {
        val draft = getDraftCompressor()
        try {
            draft.Reset()
            draft.StartNewBatch()
            
            // Copy current state to draft
            val currentState = ByteArray(goNativeBlobCompressor.Len())
            goNativeBlobCompressor.Bytes(currentState)
            
            // Try to append in draft
            val canWrite = draft.CanWrite(blockRLPEncoded, blockRLPEncoded.size)
            if (!canWrite) {
                log.debug("Block cannot be appended: {}", draft.Error())
            }
            return canWrite
        } finally {
            returnDraftCompressor(draft)
        }
    }

    override fun appendBlock(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
        val compressionSizeBefore = goNativeBlobCompressor.Len()
        val draft = getDraftCompressor()
        
        try {
            // Use draft for compression attempt
            draft.Reset()
            draft.StartNewBatch()
            
            // Copy current state to draft
            val currentState = ByteArray(goNativeBlobCompressor.Len())
            goNativeBlobCompressor.Bytes(currentState)
            
            // Try compression in draft
            val success = draft.Write(blockRLPEncoded, blockRLPEncoded.size)
            val error = draft.Error()
            
            if (!success || error != null) {
                log.error("Failed to compress block: {}", error)
                throw BlobCompressionException(error ?: "Unknown compression error")
            }
            
            // Get compressed size after successful compression
            val compressedSizeAfter = draft.Len()
            
            // If compression was successful, copy draft state to main compressor
            if (success) {
                val compressedData = ByteArray(draft.Len())
                draft.Bytes(compressedData)
                
                goNativeBlobCompressor.Reset()
                goNativeBlobCompressor.StartNewBatch()
                goNativeBlobCompressor.Write(compressedData, compressedData.size)
            }
            
            log.trace(
                "Block compressed: blockRlpSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
                blockRLPEncoded.size,
                compressionSizeBefore,
                compressedSizeAfter,
                1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / blockRLPEncoded.size)
            )
            
            return BlobCompressor.AppendResult(success, compressionSizeBefore, compressedSizeAfter)
        } finally {
            returnDraftCompressor(draft)
        }
    }

    override fun startNewBatch() {
        goNativeBlobCompressor.StartNewBatch()
    }

    override fun getCompressedData(): ByteArray {
        val compressedData = ByteArray(goNativeBlobCompressor.Len())
        goNativeBlobCompressor.Bytes(compressedData)
        return compressedData
    }

    override fun reset() {
        goNativeBlobCompressor.Reset()
    }
} 
