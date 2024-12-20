package maru.consensus.core

data class BeaconBlockBody(val prevBlockSeals: List<Seal>, val executionPayload: ExecutionPayload)
