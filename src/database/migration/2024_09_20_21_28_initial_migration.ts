import { type Kysely, sql } from 'kysely';

export async function up(db: Kysely<any>): Promise<void> {
  // User
  await db.schema
    .createTable('user')
    .addColumn('id', 'bigint', (col) => col.primaryKey())
    .addColumn('username', 'varchar(200)')
    .addColumn('first_name', 'varchar(200)')
    .addColumn('last_name', 'varchar(200)')
    .addColumn('created_at', 'timestamptz', (col) =>
      col.notNull(),
    )
    .execute();

  // Chat
  await db.schema
    .createTable('chat')
    .addColumn('id', 'bigint', (col) => col.primaryKey())
    .addColumn('command', 'varchar(100)', (col) =>
      col.notNull(),
    )
    .addColumn('step', 'int2', (col) => col.notNull())
    .addColumn('data', 'json')
    .addColumn('created_at', 'timestamptz', (col) =>
      col.notNull(),
    )
    .addColumn('updated_at', 'timestamptz')
    .execute();

  // Product
  await db.schema
    .createTable('product')
    .addColumn('id', 'bigserial', (col) => col.primaryKey())
    .addColumn('name', 'varchar(100)', (col) =>
      col.unique().notNull(),
    )
    .addColumn('url', 'varchar(2048)', (col) =>
      col.unique().notNull(),
    )
    .addColumn('sha', 'varchar(128)', (col) => col.unique())
    .addColumn('size', 'bigint')
    .addColumn('created_at', 'timestamptz', (col) =>
      col.notNull(),
    )
    .addColumn('updated_at', 'timestamptz')
    .execute();

  await sql`CREATE INDEX product_name_fulltext_index ON product USING GIN (to_tsvector('english', name))`.execute(
    db,
  );

  // Product version
  await db.schema
    .createTable('product_version')
    .addColumn('id', 'bigserial', (col) => col.primaryKey())
    .addColumn('product_id', 'bigint', (col) =>
      col.notNull(),
    )
    .addColumn('version', 'varchar(20)', (col) =>
      col.notNull(),
    )
    .addColumn('release_date', 'date', (col) =>
      col.notNull(),
    )
    .addColumn('created_at', 'timestamptz', (col) =>
      col.notNull(),
    )
    .addColumn('updated_at', 'timestamptz')
    .addForeignKeyConstraint(
      'product_version_product_id_foreign_key',
      ['product_id'],
      'product',
      ['id'],
      (eb) => eb.onUpdate('cascade').onDelete('cascade'),
    )
    .execute();

  await db.schema
    .createIndex('product_version_product_id_index')
    .on('product_version')
    .column('product_id')
    .execute();

  await db.schema
    .createIndex('product_version_version_index')
    .on('product_version')
    .column('version')
    .execute();

  await db.schema
    .createIndex('product_version_release_date_index')
    .on('product_version')
    .column('release_date')
    .execute();

  await db.schema
    .createIndex('product_version_created_at_index')
    .on('product_version')
    .column('created_at')
    .execute();

  await db.schema
    .createIndex(
      'product_version_version_product_id_unique_index',
    )
    .on('product_version')
    .unique()
    .columns(['version', 'product_id'])
    .execute();

  // Watch list
  await db.schema
    .createTable('watch_list')
    .addColumn('id', 'bigserial', (col) => col.primaryKey())
    .addColumn('chat_id', 'bigint', (col) => col.notNull())
    .addColumn('product_id', 'bigint', (col) =>
      col.notNull(),
    )
    .addColumn('created_at', 'timestamptz', (col) =>
      col.notNull(),
    )
    .addColumn('updated_at', 'timestamptz')
    .addForeignKeyConstraint(
      'watch_list_product_id_foreign_key',
      ['product_id'],
      'product',
      ['id'],
      (eb) => eb.onUpdate('cascade').onDelete('cascade'),
    )
    .execute();

  await db.schema
    .createIndex('watch_list_product_id_index')
    .on('watch_list')
    .column('product_id')
    .execute();

  await db.schema
    .createIndex(
      'watch_list_chat_id_product_id_unique_index',
    )
    .unique()
    .on('watch_list')
    .columns(['chat_id', 'product_id'])
    .execute();
}

export async function down(db: Kysely<any>): Promise<void> {
  await db.schema
    .dropTable('watch_list')
    .ifExists()
    .execute();
  await db.schema
    .dropTable('product_version')
    .ifExists()
    .execute();
  await db.schema.dropTable('product').ifExists().execute();
  await db.schema.dropTable('chat').ifExists().execute();
}
