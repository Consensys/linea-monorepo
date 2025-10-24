import { BaseError, ContractFunctionRevertedError } from "viem";
import { Result, ok, err } from "neverthrow";
import { ILogger } from "../../logging/ILogger";

type AsyncFactory<T> = () => Promise<T>;

// TODO - How do we actually want this to be implemented
export async function toViemResultAsync<T>(callback: AsyncFactory<T>, logger?: ILogger): Promise<Result<T, BaseError>> {
  try {
    const value = await callback();
    return ok(value);
  } catch (e) {
    if (err instanceof ContractFunctionRevertedError) {
      if (logger) {
        logger.error("❌ Reverted:", err.shortMessage);
        logger.error("Reason:", err.data?.errorName || err.message);
      }
    } else if (e instanceof BaseError) {
      return err(e.walk() as BaseError);
    }
    return err(e as BaseError);
  }
}
