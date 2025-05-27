export type FunctionKeys<T> = {
  [K in keyof T]: T[K] extends (...args: unknown[]) => unknown ? K : never;
}[keyof T];

export type FunctionOnly<T> = Pick<T, FunctionKeys<T>>;
