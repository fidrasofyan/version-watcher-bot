import { jsonArrayFrom } from 'kysely/helpers/postgres';
import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

export async function watchList(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  const watchList = await kysely
    .selectFrom('watch_list as wl')
    .innerJoin('product', 'product.id', 'wl.product_id')
    .select((eb) => [
      'product.name as product_name',
      jsonArrayFrom(
        eb
          .selectFrom('product_version as pv')
          .select(['pv.version', 'pv.release_date'])
          .whereRef('pv.product_id', '=', 'product.id')
          .orderBy('pv.release_date', 'desc')
          .limit(3),
      ).as('recent_releases'),
    ])
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
          `<b>${p.product_name.toUpperCase()}</b>\n` +
          p.recent_releases
            .map(
              (r) =>
                `version: <code>${r.version}</code> - release: ${r.release_date}`,
            )
            .join('\n'),
      )
      .join('\n\n');
  }

  return {
    method: 'sendMessage',
    chat_id: req.message!.chat.id,
    parse_mode: 'HTML',
    text,
    reply_markup: DEFAULT_REPLY_MARKUP,
  };
}
