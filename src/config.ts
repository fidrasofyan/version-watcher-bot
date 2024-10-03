import { readFileSync } from "node:fs";

function readEnvSync(name: string): string {
  try {
    // For docker
    const secretPath = process.env[`${name}_FILE`];
    return readFileSync(secretPath ?? "", "utf8");
  } catch (error) {
    const env = process.env[name];
    if (!env) {
      console.error(`${name} is undefined`);
      process.exit(1);
    }
    return env;
  }
}

const config = {
  // App
  NODE_ENV: readEnvSync("NODE_ENV") as "development" | "production",
  APP_TIMEZONE: readEnvSync("APP_TIMEZONE"),
  APP_NAME: readEnvSync("APP_NAME"),
  APP_HOST: readEnvSync("APP_HOST"),
  APP_PORT: parseInt(readEnvSync("APP_PORT")),

  // Token
  GITHUB_TOKEN: readEnvSync("GITHUB_TOKEN"),
  TELEGRAM_BOT_TOKEN: readEnvSync("TELEGRAM_BOT_TOKEN"),
  WEBHOOK_SECRET_TOKEN: readEnvSync("WEBHOOK_SECRET_TOKEN"),
  WEBHOOK_URL: readEnvSync("WEBHOOK_URL"),

  // Database
  DATABASE_HOST: readEnvSync("DATABASE_HOST"),
  DATABASE_PORT: parseInt(readEnvSync("DATABASE_PORT")),
  DATABASE_NAME: readEnvSync("DATABASE_NAME"),
  DATABASE_USER: readEnvSync("DATABASE_USER"),
  DATABASE_PASSWORD: readEnvSync("DATABASE_PASSWORD"),
  DATABASE_MAX_CONNECTIONS: parseInt(readEnvSync("DATABASE_MAX_CONNECTIONS")),
};

// Set timezone
process.env.TZ = config.APP_TIMEZONE;

import packageJson from "../package.json";
console.log(
  `${config.APP_NAME} # Bun: ${Bun.version} - app: v${packageJson.version} - env: ${config.NODE_ENV} - timezone: ${config.APP_TIMEZONE}`
);

export default config;
