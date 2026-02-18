import { Message } from "../../domain/message/Message";
import { MessageStatus } from "../../domain/types/enums";

import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { Direction } from "../../domain/types/enums";

export type ErrorContext = {
  operation: string;
  direction?: Direction;
  messageHash?: string;
};

export async function handleUseCaseError(params: {
  error: unknown;
  errorParser: IErrorParser;
  logger: ILogger;
  context: ErrorContext;
  repository?: IMessageRepository;
  message?: Message;
}): Promise<void> {
  const parsed = params.errorParser.parse(params.error);

  if (!parsed.mitigation.shouldRetry && params.message && params.repository) {
    params.message.edit({ status: MessageStatus.NON_EXECUTABLE });
    await params.repository.updateMessage(params.message);
  }

  params.logger[parsed.severity](
    "%s failed: errorCode=%s errorMessage=%s",
    params.context.operation,
    parsed.errorCode,
    parsed.errorMessage,
    {
      ...params.context,
      ...(parsed.data ? { data: parsed.data } : {}),
    },
  );
}
