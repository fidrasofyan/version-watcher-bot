import config from '../config';

export async function authMiddleware(
  req: Request,
): Promise<{
  ok: boolean;
}> {
  const requestToken = req.headers.get(
    'X-Telegram-Bot-Api-Secret-Token',
  );
  if (requestToken !== config.WEBHOOK_SECRET_TOKEN) {
    return { ok: false };
  }
  return { ok: true };
}
