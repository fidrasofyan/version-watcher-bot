import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

export async function unwatch1(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  const watchList = await kysely
    .selectFrom('watch_list as wl')
    .innerJoin('product', 'product.id', 'wl.product_id')
    .select(['product.name as product_name'])
    .where(
      'wl.chat_id',
      '=',
      req.message!.chat.id.toString(),
    )
    .orderBy('product.name')
    .execute();

  let text = '<b>Watch List</b>\n\n';

  if (watchList.length === 0) {
    text += '<i>No products added to watch list</i>';
  } else {
    text += watchList
      .map(
        (p) =>
          `• ${p.product_name.toUpperCase()} - /unwatch_${p.product_name.replaceAll('-', '_')}`,
      )
      .join('\n');
  }

  return {
    method: 'sendMessage',
    chat_id: req.message!.chat.id,
    parse_mode: 'HTML',
    text,
    reply_markup: DEFAULT_REPLY_MARKUP,
  };
}
