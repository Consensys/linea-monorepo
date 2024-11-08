deploymentTransactionLeadsToEmptyDeploymentTest
===============================================

hub.tx-finalization---success-setting-sender-account-row


deploymentTransactionLeadsToNonemptyDeploymentTest
==================================================

hub.return-instruction---first-account-row-for-nonempty-deployments
hub.return-instruction---justifying-the-ICPX
hub.return-instruction---setting-the-second-MMU-instruction
hub.tx-finalization---success-setting-sender-account-row


deploymentTransactionEmptyReverts
=================================

hub.tx-finalization---failure-setting-recipient-account-row
hub.tx-finalization---failure-setting-sender-account-row


deploymentTransactionNonemptyReverts
====================================

hub-into-gas
hub.tx-finalization---failure-setting-recipient-account-row
hub.tx-finalization---failure-setting-sender-account-row


createDeploysEmptyByteCode
==========================

hub-into-romlex
hub.account---the-ROMLEX-lookup-requires-nonzero-code-size
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---setting-GAS_NEXT
hub.tx-finalization---success-setting-sender-account-row


createDeploysNonemptyByteCode
=============================

hub-into-romlex
hub.account---the-ROMLEX-lookup-requires-nonzero-code-size
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---setting-GAS_NEXT
hub.return-instruction---justifying-the-ICPX
hub.tx-finalization---success-setting-sender-account-row


createRevertsWithNonemptyReturnData
===================================

hub-into-gas
hub-into-romlex
hub.account---the-ROMLEX-lookup-requires-nonzero-code-size
hub.account-instruction---foreign-address-opcode---doing-account-row
hub.account-instruction---foreign-address-opcode---setting-peeking-flags
hub.account-instruction---foreign-address-opcode---undoing-account-row
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-failures-that-will-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-CREATE-scenario---not-rebuffed-nonempty-init-code
hub.create-instruction---undoing-createe-account-operations-for-deployment-failures
hub.create-instruction---undoing-creator-account-operations-for-deployment-failures
hub.generalities---context-number-generalities
hub.tx-finalization---failure-setting-recipient-account-row
hub.tx-finalization---failure-setting-sender-account-row


createRevertsWithEmptyReturnData
================================

hub-into-romlex
hub.account---the-ROMLEX-lookup-requires-nonzero-code-size
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-failures-that-wont-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-CREATE-scenario---not-rebuffed-nonempty-init-code
hub.create-instruction---undoing-createe-account-operations-for-deployment-failures
hub.create-instruction---undoing-creator-account-operations-for-deployment-failures
hub.tx-finalization---success-setting-sender-account-row


TO INVESTIGATE NEXT
===================

hub.create-instruction---setting-the-CREATE-scenario
hub.generalities---context-number-generalities

RESOLVED
========
