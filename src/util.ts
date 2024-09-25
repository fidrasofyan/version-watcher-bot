import config from "./config";

const telegramApiUrl =
  "https://api.telegram.org/bot" + config.TELEGRAM_BOT_TOKEN + "/setWebhook";
const webhookUrl = config.WEBHOOK_URL + "/telegram-bot";
const secretToken = Buffer.from(
  crypto.getRandomValues(new Uint8Array(32))
).toString("base64url");

export async function setTelegramWebhook() {
  try {
    const result = await fetch(telegramApiUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        url: webhookUrl,
        secret_token: secretToken,
        max_connections: 50,
        drop_pending_updates: true,
        allowed_updates: ["message", "callback_query"],
      }),
    });

    if (result.status === 401 || result.status === 404) {
      console.error("Invalid Telegram Bot API token");
      process.exit(1);
    }

    await modifyEnvLine("WEBHOOK_SECRET_TOKEN", secretToken);
    config.WEBHOOK_SECRET_TOKEN = secretToken;

    console.log(
      `${config.APP_NAME} # telegram webhook has been set successfully`
    );
  } catch (error) {
    // On network error
    console.error(error);
    process.exit(1);
  }
}

export async function modifyEnvLine(linePrefix: string, newValue: string) {
  const envData = await Bun.file("./.env").text();

  const updatedEnvData = envData
    .split("\n")
    .map((line) => {
      if (line.startsWith(`${linePrefix}=`)) {
        return `${linePrefix}="${newValue}"`;
      }
      return line;
    })
    .join("\n");

  await Bun.write("./.env", updatedEnvData);
}

export function json(data: {
  statusCode: 200 | 401 | 404 | 422 | 500;
  body: Record<string, any>;
}): Response {
  return new Response(JSON.stringify(data.body), {
    status: data.statusCode,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
