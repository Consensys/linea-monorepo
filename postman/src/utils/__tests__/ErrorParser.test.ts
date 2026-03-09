import { describe, it, expect } from "@jest/globals";

import { DatabaseErrorType, DatabaseRepoName } from "../../core/enums";
import { DatabaseAccessError } from "../../core/errors";
import { ErrorParser } from "../ErrorParser";
import { generateMessage } from "../testing/helpers";

describe("ErrorParser", () => {
  describe("parseErrorWithMitigation", () => {
    it("should return null when error is null", () => {
      expect(ErrorParser.parseErrorWithMitigation(null)).toStrictEqual(null);
    });

    it("should return UNKNOWN_ERROR and shouldRetry = false when error is a generic Error", () => {
      expect(ErrorParser.parseErrorWithMitigation(new Error("any reason"))).toStrictEqual({
        errorMessage: "any reason",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          shouldRetry: false,
        },
      });
    });

    it("should return UNKNOWN_ERROR and shouldRetry = true when error = DatabaseAccessError", () => {
      const databaseAccessError = new DatabaseAccessError(
        DatabaseRepoName.MessageRepository,
        DatabaseErrorType.Insert,
        new Error("Database access failed"),
        generateMessage(),
      );

      expect(ErrorParser.parseErrorWithMitigation(databaseAccessError)).toStrictEqual({
        errorMessage: "MessageRepository: insert - Database access failed",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          shouldRetry: true,
        },
      });
    });
  });
});
