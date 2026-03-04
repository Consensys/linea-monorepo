import fs from "node:fs";
import path from "node:path";

describe("Dockerfile prisma runtime setup", () => {
  const dockerfilePath = path.join(process.cwd(), "Dockerfile");
  const dockerfileContents = fs.readFileSync(dockerfilePath, "utf8");

  it("copies prisma migrations into the runtime image", () => {
    expect(dockerfileContents).toContain(
      "COPY --from=builder --chown=node:node /usr/src/app/native-yield-operations/lido-governance-monitor/prisma ./native-yield-operations/lido-governance-monitor/prisma",
    );
  });

  it("copies prisma.config.ts into the runtime image", () => {
    expect(dockerfileContents).toContain(
      "COPY --from=builder --chown=node:node /usr/src/app/native-yield-operations/lido-governance-monitor/prisma.config.ts ./native-yield-operations/lido-governance-monitor/prisma.config.ts",
    );
  });
});
