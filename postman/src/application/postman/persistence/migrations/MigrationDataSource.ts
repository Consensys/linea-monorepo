// Configuration to generate migration files with TypeORM

import { DBOptions } from "../config/types";
import { DB } from "../dataSource";
import { config } from "dotenv";

config();

const databaseOptions: DBOptions = {
  type: "postgres",
  host: process.env.POSTGRES_HOST ?? "127.0.0.1",
  port: parseInt(process.env.POSTGRES_PORT ?? "5432"),
  username: process.env.POSTGRES_USER ?? "postgres",
  password: process.env.POSTGRES_PASSWORD ?? "postgres",
  database: process.env.POSTGRES_DB ?? "postman_db",
};

const MigrationDataSource = DB.create(databaseOptions);
export default MigrationDataSource;
