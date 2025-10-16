import { QueryParameter, DuneClient } from "@duneanalytics/client-sdk";

export function getDuneClient(duneApiKey: string): DuneClient {
  return new DuneClient(duneApiKey);
}

export function generateQueryParameter(key: string, value: string | number | Date): QueryParameter {
  if (typeof value === "string") {
    return QueryParameter.text(key, value);
  }

  if (value instanceof Date) {
    return QueryParameter.date(key, value.toISOString().slice(0, 19).replace("T", " "));
  }

  if (typeof value === "number" && Number.isInteger(value)) {
    return QueryParameter.number(key, value);
  }

  throw new Error(`Unsupported query parameter type for key: ${key}`);
}

export function generateQueryParameters(params: Record<string, string | number | Date>): QueryParameter[] {
  return Object.entries(params).map(([key, value]) => generateQueryParameter(key, value));
}

export function runDuneQuery(duneClient: DuneClient, duneQueryId: number, params: QueryParameter[] = []) {
  return duneClient.runQuery({
    queryId: duneQueryId,
    query_parameters: params,
  });
}
