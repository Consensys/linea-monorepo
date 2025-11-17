import { MigrationInterface, QueryRunner, TableColumn } from "typeorm";

export class AddClaimTxBroadcastedDateColumn1763390856686 implements MigrationInterface {
  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.addColumn(
      "message",
      new TableColumn({
        name: "claim_tx_broadcasted_date",
        type: "timestamp with time zone",
        isNullable: true,
      }),
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.dropColumn("message", "claim_tx_broadcasted_date");
  }
}
