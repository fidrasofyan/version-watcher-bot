import config from '../config';
import type {
  TelegramInlineKeyboardMarkup,
  TelegramReplyKeyboardMarkup,
} from '../types';

export class TelegramService {
  public static async sendMessage(data: {
    chat_id: number;
    parse_mode?: 'HTML' | 'MarkdownV2';
    text: string;
    reply_markup?:
      | TelegramReplyKeyboardMarkup
      | TelegramInlineKeyboardMarkup;
  }) {
    await fetch(
      `https://api.telegram.org/bot${config.TELEGRAM_BOT_TOKEN}/sendMessage`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      },
    );
  }
}
