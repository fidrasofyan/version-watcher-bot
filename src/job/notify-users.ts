import { sql } from 'kysely';
import { jsonArrayFrom } from 'kysely/helpers/postgres';
import { kysely } from '../database';
import { TelegramService } from '../service';
import { populateProductTable } from './populate-product';

export async function notifyUsers() {
  const date = await populateProductTable();

  const productsHasNewVersion = await kysely
    .selectFrom('product_version')
    .select(['product_id'])
    .distinctOn('product_id')
    .where('created_at', '=', date)
    .execute();

  if (productsHasNewVersion.length === 0) return;

  const products = await kysely
    .selectFrom('product')
    .select((eb) => [
      'product.id',
      'product.name',
      jsonArrayFrom(
        eb
          .selectFrom('product_version as pv')
          .select(['pv.version', 'pv.release_date'])
          .whereRef('pv.product_id', '=', 'product.id')
          .where('created_at', '=', date)
          .orderBy('pv.release_date', 'desc')
          .limit(5),
      ).as('recent_releases'),
    ])
    .where(
      'id',
      'in',
      productsHasNewVersion.map((p) => p.product_id),
    )
    .execute();

  const watchList = await kysely
    .selectFrom('watch_list')
    .select([
      'chat_id',
      sql<string[]>`array_agg(product_id)`.as(
        'product_ids',
      ),
    ])
    .groupBy('chat_id')
    .execute();

  for (const wl of watchList) {
    try {
      const filteredProducts = products.filter(
        (p) =>
          wl.product_ids.includes(p.id) &&
          p.recent_releases.length > 0,
      );

      if (filteredProducts.length === 0) continue;

      let text = '<b>Recent Releases Detected</b>\n\n';
      text += filteredProducts
        .map(
          (p) =>
            `<b>${p.name.toUpperCase()}</b>\n` +
            p.recent_releases
              .map(
                (r) =>
                  `version: <code>${r.version}</code> - release: ${r.release_date}`,
              )
              .join('\n'),
        )
        .join('\n\n');

      // Throttle
      await Bun.sleep(50);

      await TelegramService.sendMessage({
        chat_id: Number.parseInt(wl.chat_id),
        parse_mode: 'HTML',
        text,
      });
    } catch (error) {
      console.error(error);
    }
  }
}
