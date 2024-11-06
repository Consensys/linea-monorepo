import { createLogger, format, transports } from "winston";

// Custom format to include test name and thread/process information
const customFormat = format((info) => {
  const testName = process.env.TEST_NAME || "unknown_test";
  const threadId = process.env.JEST_WORKER_ID || "1";
  info.message = `[${testName}] [Thread ${threadId}] ${info.message}`;
  return info;
});

const logger = createLogger({
  level: "info",
  format: format.combine(
    format.timestamp(),
    customFormat(),
    format.printf(({ timestamp, level, message }) => `${timestamp} ${level}: ${message}`),
  ),
  transports: [
    new transports.Console({
      level: "debug",
      format: format.combine(
        format.colorize(),
        format.printf(({ timestamp, level, message }) => `${timestamp} ${level}: ${message}`),
      ),
    }),
    new transports.Console({
      level: "error",
      format: format.combine(
        format.colorize(),
        format.printf(({ timestamp, level, message }) => `${timestamp} ${level}: ${message}`),
      ),
    }),
  ],
});

export default logger;
