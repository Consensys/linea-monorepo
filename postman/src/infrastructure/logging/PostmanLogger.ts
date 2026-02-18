import { WinstonLogger } from "@consensys/linea-shared-utils";

import type { ILogger } from "../../domain/ports/ILogger";

export class PostmanLogger extends WinstonLogger implements ILogger {}
