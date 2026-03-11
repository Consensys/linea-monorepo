import { MigrationInterface, QueryRunner, TableColumn } from "typeorm";

export class AddClaimCycleCount1741700000000 implements MigrationInterface {
  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.addColumn(
      "message",
      new TableColumn({
        name: "claimCycleCount",
        type: "integer",
        default: 0,
        isNullable: false,
      }),
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.dropColumn("message", "claimCycleCount");
  }
}
