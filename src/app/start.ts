import config from '../config';
import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

export async function start(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  // Add to user
  const user = await kysely
    .selectFrom('user')
    .where('id', '=', req.message!.chat.id.toString())
    .select('id')
    .executeTakeFirst();

  if (!user) {
    await kysely
      .insertInto('user')
      .values({
        id: req.message!.chat.id.toString(),
        username: req.message!.chat.username,
        first_name: req.message!.chat.first_name,
        last_name: req.message!.chat.last_name,
        created_at: new Date(),
      })
      .executeTakeFirst();
  }

  return {
    method: 'sendMessage',
    chat_id: req.message!.chat.id,
    text: `Welcome to ${config.APP_NAME}. Type /help to see the list of available commands.`,
    reply_markup: DEFAULT_REPLY_MARKUP,
  };
}
