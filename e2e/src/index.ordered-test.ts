// Here we want to guarantee the running order of test files
// to minimize timeout errors in ci pipeline
import submissionAndFinalizationTestSuite from "./submission-finalization.spec";
import layer2TestSuite from "./l2.spec";
import messagingTestSuite from "./messaging.spec";
import coordinatorRestartTestSuite from "./restart.spec";
import transactionExclusionTestSuite from "./transaction-exclusion.spec";

messagingTestSuite("Messaging test suite");
// NOTES: The coordinator restart test must not be run first in the sequence of tests.
coordinatorRestartTestSuite("Coordinator restart test suite");
layer2TestSuite("Layer 2 test suite");
submissionAndFinalizationTestSuite("Submission and finalization test suite");
transactionExclusionTestSuite("Transaction exclusion test suite");
