import { randomUUID, randomInt } from "crypto";

export function generateRandomInt(max = 1000): number {
  return randomInt(max);
}

export function generateRandomUUIDv4(): string {
  return randomUUID();
}
