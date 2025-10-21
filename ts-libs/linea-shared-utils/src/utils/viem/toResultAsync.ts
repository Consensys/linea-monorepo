import { BaseError } from "viem";
import { Result, ok, err } from "neverthrow";

type AsyncFactory<T> = () => Promise<T>;

export async function toResultAsync<T>(callback: AsyncFactory<T>): Promise<Result<T, Error>> {
  try {
    const value = await callback();
    return ok(value);
  } catch (e) {
    if (e instanceof BaseError) {
      return err(e.walk() as BaseError);
    }
    return err(e as BaseError);
  }
}
