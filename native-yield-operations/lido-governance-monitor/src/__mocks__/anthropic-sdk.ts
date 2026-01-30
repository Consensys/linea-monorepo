import { jest } from "@jest/globals";

class MockAnthropic {
  messages = {
    create: jest.fn(),
  };

  constructor(_options?: { apiKey?: string }) {}
}

export default MockAnthropic;
