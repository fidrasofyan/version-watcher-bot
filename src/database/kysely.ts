import {
  FileMigrationProvider,
  Kysely,
  Migrator,
  PostgresDialect,
  sql,
} from "kysely";
import type { DB } from "kysely-codegen";
import { Pool } from "pg";
import fs from "node:fs/promises";
import path from "node:path";
import config from "../config";

const pool = new Pool({
  host: config.DATABASE_HOST,
  port: config.DATABASE_PORT,
  database: config.DATABASE_NAME,
  user: config.DATABASE_USER,
  password: config.DATABASE_PASSWORD,
  max: config.DATABASE_MAX_CONNECTIONS,
});

pool.on("connect", async (client) => {
  await client.query(`SET TIME ZONE '${config.APP_TIMEZONE}'`);
});

const dialect = new PostgresDialect({
  pool,
});

export const kysely = new Kysely<DB>({
  dialect,
});

// Migration
export async function migrate(type: "latest" | "down") {
  const kysely = new Kysely<DB>({
    dialect,
  });

  const migrator = new Migrator({
    db: kysely,
    provider: new FileMigrationProvider({
      fs,
      path,
      migrationFolder: path.join(__dirname, "./migration"),
    }),
  });

  const { error, results } =
    type === "latest"
      ? await migrator.migrateToLatest()
      : await migrator.migrateDown();

  results?.forEach((it) => {
    if (it.status === "Success") {
      console.log(
        `migration "${it.migrationName}" was ${
          type === "latest" ? "applied" : "rolled back"
        } successfully`
      );
    } else if (it.status === "Error") {
      console.error(
        `failed to ${type === "latest" ? "apply" : "rollback"} migration "${
          it.migrationName
        }"`
      );
    }
  });

  await kysely.destroy();

  if (error) {
    console.error("failed to migrate");
    console.error(error);
    process.exit(1);
  }
}

// Test connection
await sql`SELECT 1`.execute(kysely);
console.log(
  `${config.APP_NAME} # database connection has been established successfully`
);
