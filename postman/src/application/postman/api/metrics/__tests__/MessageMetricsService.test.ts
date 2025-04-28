import { EntityManager, SelectQueryBuilder } from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MessageMetricsService } from "../MessageMetricsService";
import { mock, MockProxy } from "jest-mock-extended";
import { MessageStatus } from "../../../../../core/enums";
import { LineaPostmanMetrics } from "../../../../../core/metrics/IMetricsService";

describe("MessageMetricsService", () => {
  let messageMetricsService: MessageMetricsService;
  let mockEntityManager: MockProxy<EntityManager>;

  beforeEach(() => {
    mockEntityManager = mock<EntityManager>();
    messageMetricsService = new MessageMetricsService(mockEntityManager);
  });

  it("should get correct gauge values after initialization", async () => {
    jest.spyOn(mockEntityManager, "maximum").mockResolvedValue(10);
    jest.spyOn(mockEntityManager, "createQueryBuilder").mockReturnValue({
      select: jest.fn().mockReturnThis(),
      addSelect: jest.fn().mockReturnThis(),
      groupBy: jest.fn().mockReturnThis(),
      addGroupBy: jest.fn().mockReturnThis(),
      getRawMany: jest.fn().mockResolvedValue([
        { status: MessageStatus.SENT, direction: Direction.L1_TO_L2, isForSponsorship: false, count: 2 },
        { status: MessageStatus.SENT, direction: Direction.L1_TO_L2, isForSponsorship: true, count: 3 },
        {
          status: MessageStatus.CLAIMED_SUCCESS,
          direction: Direction.L1_TO_L2,
          isForSponsorship: true,
          count: 10,
        },
      ]),
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as unknown as SelectQueryBuilder<any>);

    await messageMetricsService.initialize();

    // Check if the gauge was updated
    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
        status: MessageStatus.SENT,
        direction: Direction.L1_TO_L2,
      }),
    ).toBe(5);

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
        status: MessageStatus.SENT,
        direction: Direction.L1_TO_L2,
        isForSponsorship: String(false),
      }),
    ).toBe(2);

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
        status: MessageStatus.SENT,
        direction: Direction.L1_TO_L2,
        isForSponsorship: String(true),
      }),
    ).toBe(3);

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
        status: MessageStatus.CLAIMED_SUCCESS,
        direction: Direction.L1_TO_L2,
      }),
    ).toBe(10);
  });

  it("should get correct gauge values after increment the gauge", async () => {
    messageMetricsService.incrementGauge(
      LineaPostmanMetrics.Messages,
      {
        status: "processed",
        direction: Direction.L1_TO_L2,
        isForSponsorship: String(true),
      },
      10,
    );
    const gaugeValue = await messageMetricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
      status: "processed",
      direction: Direction.L1_TO_L2,
      isForSponsorship: String(true),
    });
    expect(gaugeValue).toBe(10);
  });
});
