export function formatInTimeZone(date: Date, _timeZone: string, _format: string): string {
  void _timeZone;
  void _format;
  return date.toISOString().slice(0, 10);
}

export function fromZonedTime(date: Date | number, _timeZone: string): Date {
  void _timeZone;
  const input = new Date(date);
  return new Date(input.getTime() - input.getTimezoneOffset() * 60_000);
}

export function toZonedTime(date: Date | number, _timeZone: string): Date {
  void _timeZone;
  const input = new Date(date);
  return new Date(input.getTime() + input.getTimezoneOffset() * 60_000);
}
