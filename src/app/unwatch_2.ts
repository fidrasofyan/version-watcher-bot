import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

export async function unwatch2(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  const productName = req
    .message!.text?.replace('/unwatch_', '')
    .replaceAll('_', '-');

  if (!productName) {
    return {
      method: 'sendMessage',
      chat_id: req.message!.chat.id,
      parse_mode: 'HTML',
      text: '<i>Unknown command</i>',
      reply_markup: DEFAULT_REPLY_MARKUP,
    };
  }

  const product = await kysely
    .selectFrom('product')
    .select(['id', 'name'])
    .where('name', '=', productName)
    .executeTakeFirst();

  if (!product) {
    return {
      method: 'sendMessage',
      chat_id: req.message!.chat.id,
      parse_mode: 'HTML',
      text: '<i>Product not found</i>',
      reply_markup: DEFAULT_REPLY_MARKUP,
    };
  }

  const result = await kysely
    .deleteFrom('watch_list')
    .where('chat_id', '=', req.message!.chat.id.toString())
    .where('product_id', '=', product.id)
    .executeTakeFirstOrThrow();

  if (result.numDeletedRows <= 0) {
    return {
      method: 'sendMessage',
      chat_id: req.message!.chat.id,
      parse_mode: 'HTML',
      text: '<i>Failed to remove product</i>',
      reply_markup: DEFAULT_REPLY_MARKUP,
    };
  }

  return {
    method: 'sendMessage',
    chat_id: req.message!.chat.id,
    parse_mode: 'HTML',
    text: `<i>${product.name.toUpperCase()} removed from watch list</i>`,
    reply_markup: DEFAULT_REPLY_MARKUP,
  };
}
