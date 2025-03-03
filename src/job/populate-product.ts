import config from '../config';
import { kysely } from '../database';

export async function populateProductTable() {
  const date = new Date();

  await kysely.transaction().execute(async (trx) => {
    // Populate product
    const fetchProductResponse = await fetch(
      'https://api.github.com/repos/endoflife-date/release-data/contents/releases',
      {
        headers: {
          Authorization: `token ${config.GITHUB_TOKEN}`,
          Accept: 'application/vnd.github.raw+json',
        },
      },
    );
    if (!fetchProductResponse.ok) {
      throw new Error(
        `HTTP error! status: ${fetchProductResponse.status}`,
      );
    }
    const releases =
      (await fetchProductResponse.json()) as {
        name: string;
        sha: string;
        size: number;
        url: string;
      }[];

    await trx
      .insertInto('product')
      .values(
        releases.map((release) => ({
          name: Bun.escapeHTML(
            release.name.replace('.json', ''),
          ),
          url: Bun.escapeHTML(release.url),
          sha: Bun.escapeHTML(release.sha),
          size: release.size,
          created_at: date,
        })),
      )
      .onConflict((oc) =>
        oc.column('name').doUpdateSet({
          url: (eb) => eb.ref('excluded.url'),
          sha: (eb) => eb.ref('excluded.sha'),
          size: (eb) => eb.ref('excluded.size'),
          updated_at: date,
        }),
      )
      .executeTakeFirstOrThrow();

    // Populate product_version based on watch_list
    const watchedProducts = await trx
      .selectFrom('product')
      .innerJoin(
        'watch_list',
        'watch_list.product_id',
        'product.id',
      )
      .select(['product.id', 'product.url'])
      .distinctOn('product.id')
      .execute();

    for (const product of watchedProducts) {
      // Throttle
      await Bun.sleep(500);

      const fetchContentResponse = await fetch(
        product.url,
        {
          headers: {
            Authorization: `token ${config.GITHUB_TOKEN}`,
            Accept: 'application/vnd.github.raw+json',
          },
        },
      );

      if (config.NODE_ENV === 'development') {
        console.log(
          'X-RateLimit-Used: ',
          fetchContentResponse.headers.get(
            'X-RateLimit-Used',
          ),
        );
        console.log(
          'X-RateLimit-Remaining: ',
          fetchContentResponse.headers.get(
            'X-RateLimit-Remaining',
          ),
        );
      }

      if (!fetchContentResponse.ok) {
        throw new Error(
          `HTTP error! status: ${fetchContentResponse.status}`,
        );
      }

      const data = (await fetchContentResponse.json()) as {
        versions: Record<
          string,
          {
            name: string;
            date: string;
          }
        >;
      };

      const productVersions: {
        product_id: string;
        version: string;
        release_date: string;
        created_at: Date;
      }[] = [];

      for (const key in data.versions) {
        const version = data.versions[key];
        productVersions.push({
          product_id: product.id,
          version: Bun.escapeHTML(version.name),
          release_date: Bun.escapeHTML(version.date),
          created_at: date,
        });
      }

      await trx
        .insertInto('product_version')
        .values(productVersions)
        .onConflict((oc) =>
          oc.columns(['version', 'product_id']).doNothing(),
        )
        .executeTakeFirstOrThrow();
    }
  });

  return date;
}
