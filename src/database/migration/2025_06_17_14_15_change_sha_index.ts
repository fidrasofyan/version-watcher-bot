import type { Kysely } from 'kysely';

export async function up(db: Kysely<any>): Promise<void> {
  await db.schema
    .alterTable('product')
    .dropConstraint('product_sha_key')
    .execute();

  await db.schema
    .createIndex('product_sha_index')
    .on('product')
    .column('sha')
    .execute();
}

export async function down(db: Kysely<any>): Promise<void> {
  await db.schema
    .createIndex('product_sha_key')
    .unique()
    .on('product')
    .column('sha')
    .execute();

  await db.schema
    .alterTable('product')
    .dropIndex('product_sha_index')
    .execute();
}
