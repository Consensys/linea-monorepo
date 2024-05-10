// Here we want to guarantee the running order of test files
// to minimize timeout errors in ci pipeline
import contractMigrationTestSuite from "./contract-migration.spec";
import layer2TestSuite from "./l2.spec";
import messagingTestSuite from "./messaging.spec";
import coordinatorRestartTestSuite from "./restart.spec";

coordinatorRestartTestSuite("Coordinator restart test suite");
layer2TestSuite("Layer 2 test suite");
messagingTestSuite("Messaging test suite");
contractMigrationTestSuite("Contract migration test suite");
