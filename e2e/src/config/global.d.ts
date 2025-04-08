/* eslint-disable no-var */

import { Logger } from "winston";

declare global {
  var stopL2TrafficGeneration: () => void;
  var skipSendBundleTests: boolean;
  var logger: Logger;
}

export {};
