const CHINESE_TEXT = /[\u3400-\u9fff]/;

export function userFacingErrorMessage(error: unknown, fallback = '网络不稳，请稍后再试'): string {
  const message = typeof (error as {message?: unknown})?.message === 'string'
    ? (error as {message: string}).message.trim()
    : '';
  if (!message) return fallback;
  return CHINESE_TEXT.test(message) ? message : fallback;
}
