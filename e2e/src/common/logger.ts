import { createLogger, format, transports } from "winston";

const logger = createLogger({
  level: "info",
  format: format.combine(
    format.timestamp(),
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
