import { CronJob } from 'cron';
import config from './config';
import { kysely } from './database';
import { handler } from './handler';
import { notifyUsers } from './job';
import { authMiddleware } from './middleware/auth-middleware';
import type {
  TelegramRequest,
  TelegramResponse,
} from './types';
import {
  generateDatetime,
  json,
  setTelegramWebhook,
} from './util';

// Set Telegram webhook
if (config.NODE_ENV === 'production') {
  await setTelegramWebhook();
}

// HTTP server
const httpServer = Bun.serve({
  hostname: config.APP_HOST,
  port: config.APP_PORT,
  async fetch(req) {
    try {
      const url = new URL(req.url);

      if (url.pathname === '/telegram-bot') {
        // Auth middleware
        const auth = await authMiddleware(req);
        if (!auth.ok) {
          return json({
            statusCode: 401,
            body: {
              success: false,
              message: 'Unauthorized',
            },
          });
        }

        // Request
        const telegramRequest: TelegramRequest =
          await req.json();

        try {
          // Handler
          const result = await handler(telegramRequest);

          return json({
            statusCode: 200,
            body: result,
          });
        } catch (error) {
          console.error(error);

          const body: TelegramResponse = {
            method: 'sendMessage',
            chat_id: telegramRequest.message
              ? telegramRequest.message.chat.id
              : telegramRequest.callback_query
                ? telegramRequest.callback_query.from.id
                : 0,
            parse_mode: 'HTML',
            text: '<i>Internal server error</i>',
          };

          return json({
            statusCode: 200,
            body: body,
          });
        }
      }

      // Not found
      return json({
        statusCode: 404,
        body: {
          success: false,
          message: 'Not found',
        },
      });
    } catch (error) {
      console.error(error);

      return json({
        statusCode: 500,
        body: {
          success: false,
          message: 'Internal server error',
        },
      });
    }
  },
});

console.log(
  `${config.APP_NAME} (${generateDatetime()}) # server running at ${httpServer.url}`,
);

// Cron
if (config.NODE_ENV === 'production') {
  new CronJob(
    '0 * * * *', // Every hour
    async () => {
      console.log(
        `${config.APP_NAME} (${generateDatetime()}) # cron: notifying users...`,
      );

      await notifyUsers();

      console.log(
        `${
          config.APP_NAME
        } (${generateDatetime()}) # cron: notifying users... done`,
      );
    },
    null,
    true,
  );

  console.log(
    `${config.APP_NAME} (${generateDatetime()}) # cron job has been started successfully`,
  );
}

// Graceful shutdown
process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);

async function shutdown() {
  httpServer.stop();
  await kysely.destroy();

  console.log(
    `${config.APP_NAME} (${generateDatetime()}) # server has been stopped successfully`,
  );
  process.exit(0);
}
