export function reportHandledError(scope: string, error: unknown): void {
  if (process.env.NODE_ENV === 'test') return;
  console.log(`[handled] ${scope}`, error);
}
