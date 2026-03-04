import fs from "node:fs";
import path from "node:path";

describe("prisma config", () => {
  const prismaConfigPath = path.join(process.cwd(), "prisma.config.ts");
  const prismaConfigContents = fs.readFileSync(prismaConfigPath, "utf8");

  it("requires DATABASE_URL for datasource configuration", () => {
    expect(prismaConfigContents).toContain('url: env("DATABASE_URL")');
  });

  it("does not mark DATABASE_URL as optional", () => {
    expect(prismaConfigContents).not.toContain('env("DATABASE_URL", { optional: true })');
  });
});
