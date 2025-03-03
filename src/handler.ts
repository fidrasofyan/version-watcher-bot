import {
  chatId,
  notFound,
  start,
  unwatch,
  watch,
  watchList,
} from './app';
import { DEFAULT_REPLY_MARKUP } from './constant';
import { TelegramRepository } from './repository/telegram-repository';
import type {
  TelegramRequest,
  TelegramResponse,
} from './types';

export async function handler(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  const data = {
    chatId: 0,
    command: '',
  };

  // If callback query
  if (req.callback_query) {
    data.chatId = req.callback_query.from.id;
  }
  // If text message
  else if (req.message) {
    data.chatId = req.message.chat.id;

    if (!req.message.text) {
      return {
        method: 'sendMessage',
        chat_id: req.message.chat.id,
        parse_mode: 'HTML',
        text: '<i>Only text messages are allowed</i>',
      };
    }

    data.command = req.message.text
      .slice(0, 30)
      .toLowerCase()
      .replace(/[\s_/]+/g, '');

    if (data.command === 'cancel') {
      await TelegramRepository.deleteChat(
        req.message.chat.id,
      );
      return {
        method: 'sendMessage',
        chat_id: req.message.chat.id,
        parse_mode: 'HTML',
        text: '<i>Cancelled</i>',
        reply_markup: DEFAULT_REPLY_MARKUP,
      };
    }
  }

  const chat = await TelegramRepository.getChat(
    data.chatId,
  );
  if (chat) {
    data.command = chat.command;
  }

  switch (data.command) {
    case 'chatid': {
      return chatId(req);
    }

    case 'start': {
      return start(req);
    }

    case 'unwatch': {
      return unwatch(req);
    }

    case 'watchlist': {
      return watchList(req);
    }

    case 'watch': {
      return watch(req);
    }

    default: {
      return notFound(req);
    }
  }
}
