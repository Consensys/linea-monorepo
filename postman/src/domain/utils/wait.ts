export const wait = (timeout: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, timeout));
