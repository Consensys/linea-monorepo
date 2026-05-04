import { WinstonLogger } from "../WinstonLogger";

// eslint-disable-next-line no-control-regex
const stripAnsi = (str: string) => str.replace(/\x1B\[[0-9;]*m/g, "");

describe("WinstonLogger", () => {
  let consoleSpy: jest.SpyInstance;

  beforeEach(() => {
    consoleSpy = jest.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  function getLogOutput(): string {
    return stripAnsi(consoleSpy.mock.calls.map((call) => String(call[0])).join(""));
  }

  describe("log line structure", () => {
    it("contains time, level, logger, and msg fields", () => {
      const logger = new WinstonLogger("MyService");
      logger.info("hello world");

      const output = getLogOutput();
      expect(output).toMatch(/time=\S+/);
      expect(output).toContain("level=INFO");
      expect(output).toContain("logger=MyService");
      expect(output).toContain('msg="hello world"');
    });

    it("is a single line even when metadata contains newlines", () => {
      const logger = new WinstonLogger("Test");
      logger.info("msg", { reason: "line1\nline2" });

      const lines = getLogOutput()
        .trim()
        .split("\n")
        .filter((l) => l.includes("level="));
      expect(lines).toHaveLength(1);
    });

    it("uppercases the level", () => {
      new WinstonLogger("Test").warn("watch out");
      expect(getLogOutput()).toContain("level=WARN");
    });

    it("only colorizes the level value, not the whole line", () => {
      new WinstonLogger("Test").info("msg");
      const raw = consoleSpy.mock.calls.map((call) => String(call[0])).join("");
      // eslint-disable-next-line no-control-regex
      const ansiLevel = /level=\x1B\[\d+mINFO\x1B\[\d+m/;
      expect(raw).toMatch(ansiLevel);
      // eslint-disable-next-line no-control-regex
      expect(raw.replace(ansiLevel, "")).not.toMatch(/\x1B\[/);
    });
  });

  describe("log levels", () => {
    it("logs at info level", () => {
      new WinstonLogger("Test").info("info msg");
      expect(getLogOutput()).toContain("level=INFO");
    });

    it("logs at warn level", () => {
      new WinstonLogger("Test").warn("warn msg");
      expect(getLogOutput()).toContain("level=WARN");
    });

    it("logs at error level", () => {
      new WinstonLogger("Test").error("error msg");
      expect(getLogOutput()).toContain("level=ERROR");
    });

    it("does not log debug by default", () => {
      new WinstonLogger("Test").debug("debug msg");
      expect(getLogOutput()).toBe("");
    });

    it("logs debug when level option is debug", () => {
      new WinstonLogger("Test", { level: "debug" }).debug("debug msg");
      expect(getLogOutput()).toContain("level=DEBUG");
    });
  });

  describe("logfmt value quoting", () => {
    it("quotes values that contain spaces", () => {
      new WinstonLogger("Test").info("msg", { reason: "something went wrong" });
      expect(getLogOutput()).toContain('reason="something went wrong"');
    });

    it("does not quote simple values", () => {
      new WinstonLogger("Test").info("msg", { foo: "bar" });
      expect(getLogOutput()).toContain("foo=bar");
    });

    it("escapes newlines inside values to keep the entry on one line", () => {
      new WinstonLogger("Test").info("msg", { reason: "line1\nline2" });
      expect(getLogOutput()).toContain('reason="line1\\nline2"');
    });

    it("escapes double-quotes inside values", () => {
      new WinstonLogger("Test").info("msg", { note: 'say "hello"' });
      expect(getLogOutput()).toContain('note="say \\"hello\\""');
    });

    it("escapes backslashes inside values", () => {
      new WinstonLogger("Test").info("msg", { path: "C:\\Users\\foo" });
      expect(getLogOutput()).toContain('path="C:\\\\Users\\\\foo"');
    });
  });

  describe("metadata — no metadata", () => {
    it("produces no extra fields when there is no metadata", () => {
      new WinstonLogger("Test").info("plain");
      const output = getLogOutput();
      expect(output.trim()).toMatch(/^time=\S+ level=\S+ logger=\S+ msg=plain$/);
    });
  });

  describe("metadata — scalar values", () => {
    it("includes multiple string fields", () => {
      new WinstonLogger("Test").info("msg", { a: "1", b: "2" });
      const output = getLogOutput();
      expect(output).toContain("a=1");
      expect(output).toContain("b=2");
    });

    it("serializes numbers directly", () => {
      new WinstonLogger("Test").info("msg", { count: 42 });
      expect(getLogOutput()).toContain("count=42");
    });

    it("serializes booleans directly", () => {
      new WinstonLogger("Test").info("msg", { active: false });
      expect(getLogOutput()).toContain("active=false");
    });

    it("serializes bigints without JSON quotes", () => {
      new WinstonLogger("Test").info("msg", { value: BigInt("99999999999999999999") });
      expect(getLogOutput()).toContain("value=99999999999999999999");
    });
  });

  describe("metadata — object values", () => {
    it("serializes plain objects as quoted JSON", () => {
      new WinstonLogger("Test").info("msg", { meta: { x: 1 } });
      expect(getLogOutput()).toContain('meta="{\\"x\\":1}"');
    });
  });

  describe("metadata — Error values", () => {
    it("outputs the error stack as an escaped single-line value", () => {
      const logger = new WinstonLogger("Test");
      const error = new Error("something went wrong\nwith a newline");

      logger.error("oops", { error });

      const output = getLogOutput();
      expect(output).toContain('"Error: something went wrong\\nwith a newline');
      expect(
        output
          .trim()
          .split("\n")
          .filter((l) => l.includes("level=")),
      ).toHaveLength(1);
    });

    it("falls back to message when stack is unavailable", () => {
      const error = new Error("no stack");
      delete error.stack;

      new WinstonLogger("Test").error("oops", { error });

      expect(getLogOutput()).toContain('error="no stack"');
    });

    it("does not produce an error.stack key", () => {
      new WinstonLogger("Test").error("oops", { error: new Error("msg") });
      expect(getLogOutput()).not.toContain("error.stack=");
    });
  });

  describe("normalizeParams — bare Error as second argument", () => {
    it("does not corrupt the msg field", () => {
      const logger = new WinstonLogger("Test");
      logger.error("something failed", new Error("direct error"));

      const output = getLogOutput();
      expect(output).toContain('msg="something failed"');
      expect(output).not.toContain("something failed direct error");
    });

    it("outputs the stack under the error key", () => {
      new WinstonLogger("Test").error("something failed", new Error("direct error"));
      expect(getLogOutput()).toContain("Error: direct error");
    });

    it("does not scatter enumerable Error properties into metadata", () => {
      const error = Object.assign(new Error("conn refused"), { code: "ECONNREFUSED", errno: -61 });
      new WinstonLogger("Test").error("startup failed", error);

      const output = getLogOutput();
      expect(output).not.toContain("code=ECONNREFUSED");
      expect(output).not.toContain("errno=-61");
    });
  });
});
