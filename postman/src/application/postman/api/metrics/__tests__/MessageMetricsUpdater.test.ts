import { EntityManager, SelectQueryBuilder } from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MessageMetricsUpdater } from "../MessageMetricsUpdater";
import { mock, MockProxy } from "jest-mock-extended";
import { MessageStatus } from "../../../../../core/enums";
import { SingletonMetricsService } from "../SingletonMetricsService";
import { IMessageMetricsUpdater } from "../../../../../core/metrics";

describe("MessageMetricsUpdater", () => {
  let messageMetricsUpdater: IMessageMetricsUpdater;
  let mockEntityManager: MockProxy<EntityManager>;

  beforeEach(() => {
    mockEntityManager = mock<EntityManager>();
    const metricService = new SingletonMetricsService();
    messageMetricsUpdater = new MessageMetricsUpdater(mockEntityManager, metricService);
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

  it("should get correct gauge values after initialization", async () => {
    jest.spyOn(mockEntityManager, "maximum").mockResolvedValue(10);
    jest.spyOn(mockEntityManager, "createQueryBuilder").mockReturnValue({
      select: jest.fn().mockReturnThis(),
      addSelect: jest.fn().mockReturnThis(),
      groupBy: jest.fn().mockReturnThis(),
      addGroupBy: jest.fn().mockReturnThis(),
      where: jest.fn().mockReturnThis(),
      andWhere: jest.fn().mockReturnThis(),
      getRawMany: jest.fn().mockResolvedValue(getMessagesCountQueryResp),
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as unknown as SelectQueryBuilder<any>);

    await messageMetricsUpdater.initialize();

    // Check if the gauge was updated
    expect(
      await messageMetricsUpdater.getMessageCount({
        status: MessageStatus.SENT,
        direction: Direction.L1_TO_L2,
        isForSponsorship: false,
      }),
    ).toBe(2);

    expect(
      await messageMetricsUpdater.getMessageCount({
        status: MessageStatus.SENT,
        direction: Direction.L1_TO_L2,
        isForSponsorship: true,
      }),
    ).toBe(3);

    expect(
      await messageMetricsUpdater.getMessageCount({
        status: MessageStatus.CLAIMED_SUCCESS,
        direction: Direction.L1_TO_L2,
        isForSponsorship: true,
      }),
    ).toBe(10);
  });

  it("should get correct gauge values for LineaPostmanMetrics.Messages after incrementing the gauge", async () => {
    messageMetricsUpdater.incrementMessageCount(
      {
        status: MessageStatus.PENDING,
        direction: Direction.L1_TO_L2,
        isForSponsorship: true,
      },
      10,
    );
    const gaugeValue = await messageMetricsUpdater.getMessageCount({
      status: MessageStatus.PENDING,
      direction: Direction.L1_TO_L2,
      isForSponsorship: true,
    });
    expect(gaugeValue).toBe(10);
  });
});
