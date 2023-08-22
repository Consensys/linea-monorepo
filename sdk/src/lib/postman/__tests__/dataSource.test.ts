import { describe, it, expect } from "@jest/globals";
import { DataSource } from "typeorm";
import { SnakeNamingStrategy } from "typeorm-naming-strategies";
import { DB } from "../dataSource";
import { MessageEntity } from "../entity/Message.entity";
import { InitialDatabaseSetup1685985945638 } from "../migrations/1685985945638-InitialDatabaseSetup";
import { AddNewColumns1687890326970 } from "../migrations/1687890326970-AddNewColumns";
import { UpdateStatusColumn1687890694496 } from "../migrations/1687890694496-UpdateStatusColumn";
import { RemoveUniqueConstraint1689084924789 } from "../migrations/1689084924789-RemoveUniqueConstraint";

describe("DB", () => {
  describe("create", () => {
    it("should return TypeORM DataSource instance", async () => {
      expect(
        DB.create({
          type: "postgres",
          host: "127.0.0.1",
          port: 5432,
          username: "postgres",
          password: "postgres",
          database: "db_name",
        }),
      ).toStrictEqual(
        new DataSource({
          type: "postgres",
          host: "127.0.0.1",
          port: 5432,
          username: "postgres",
          password: "postgres",
          database: "db_name",
          entities: [MessageEntity],
          namingStrategy: new SnakeNamingStrategy(),
          migrations: [
            InitialDatabaseSetup1685985945638,
            AddNewColumns1687890326970,
            UpdateStatusColumn1687890694496,
            RemoveUniqueConstraint1689084924789,
          ],
          migrationsTableName: "migrations",
          logging: ["error"],
          migrationsRun: true,
        }),
      );
    });
  });
});
