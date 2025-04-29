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

  const getMessagesCountQueryResp = [
    { status: MessageStatus.SENT, direction: Direction.L1_TO_L2, isForSponsorship: false, count: "2" },
    { status: MessageStatus.SENT, direction: Direction.L1_TO_L2, isForSponsorship: true, count: "3" },
    {
      status: MessageStatus.CLAIMED_SUCCESS,
      direction: Direction.L1_TO_L2,
      isForSponsorship: true,
      count: "10",
    },
  ];

  const WEI_STRING = "480173904";
  const GWEI_STRING = "87498317";
  const getSponsorshipFeesQueryResp = [{ direction: Direction.L1_TO_L2, totalTxFees: `${GWEI_STRING}${WEI_STRING}` }];

  it("should get correct gauge values after initialization", async () => {
    jest.spyOn(mockEntityManager, "maximum").mockResolvedValue(10);
    jest.spyOn(mockEntityManager, "createQueryBuilder").mockReturnValue({
      select: jest.fn().mockReturnThis(),
      addSelect: jest.fn().mockReturnThis(),
      groupBy: jest.fn().mockReturnThis(),
      addGroupBy: jest.fn().mockReturnThis(),
      where: jest.fn().mockReturnThis(),
      andWhere: jest.fn().mockReturnThis(),
      getRawMany: jest
        .fn()
        .mockResolvedValueOnce(getMessagesCountQueryResp)
        .mockResolvedValueOnce(getSponsorshipFeesQueryResp),
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

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesWei, {
        direction: Direction.L2_TO_L1,
      }),
    ).toBe(0);

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesGwei, {
        direction: Direction.L2_TO_L1,
      }),
    ).toBe(0);

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesWei, {
        direction: Direction.L1_TO_L2,
      }),
    ).toBe(parseInt(WEI_STRING));

    expect(
      await messageMetricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesGwei, {
        direction: Direction.L1_TO_L2,
      }),
    ).toBe(parseInt(GWEI_STRING));
  });

  it("should get correct gauge values for LineaPostmanMetrics.Messages after incrementing the gauge", async () => {
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
