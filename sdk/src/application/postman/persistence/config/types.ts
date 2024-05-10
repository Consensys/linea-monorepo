import { BetterSqlite3ConnectionOptions } from "typeorm/driver/better-sqlite3/BetterSqlite3ConnectionOptions";
import { PostgresConnectionOptions } from "typeorm/driver/postgres/PostgresConnectionOptions";

export type DBConfig = PostgresConnectionOptions | BetterSqlite3ConnectionOptions;

export type DBCleanerConfig = {
  enabled: boolean;
  cleaningInterval?: number;
  daysBeforeNowToDelete?: number;
};
