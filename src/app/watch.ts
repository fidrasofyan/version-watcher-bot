import { sql } from 'kysely';
import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import { TelegramRepository } from '../repository/telegram-repository';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

const COMMAND = 'watch';

export async function watch(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  const data = {
    isCallbackQuery: false,
    messageId: 0,
    chatId: 0,
    text: '',
  };

  if (req.message) {
    data.messageId = req.message.message_id;
    data.chatId = req.message.chat.id;
    data.text = req.message.text!;
  } else if (req.callback_query) {
    data.isCallbackQuery = true;
    data.messageId = req.callback_query.message!.message_id;
    data.chatId = req.callback_query.from.id;
    data.text = req.callback_query.message!.text!;
  }

  let chat = await TelegramRepository.getChat(data.chatId);
  if (!chat) {
    chat = await TelegramRepository.setChat({
      chatId: data.chatId,
      command: COMMAND,
      step: 1,
      data: {},
    });
  }

  switch (chat.step) {
    // Step 1
    case 1: {
      await TelegramRepository.setChat({
        chatId: data.chatId,
        command: COMMAND,
        step: 2,
      });

      return {
        method: 'sendMessage',
        chat_id: data.chatId,
        parse_mode: 'HTML',
        text: [
          '<i>Type what you want to watch...</i>',
          '\n<i>E.g. "Ubuntu", "nginx"</i>',
        ].join('\n'),
        reply_markup: DEFAULT_REPLY_MARKUP,
      };
    }

    // Step 2
    case 2: {
      if (req.callback_query?.data === 'cancel') {
        await TelegramRepository.deleteChat(data.chatId);
        return {
          method: 'editMessageText',
          message_id: data.messageId,
          chat_id: data.chatId,
          parse_mode: 'HTML',
          text: '<i>Cancelled</i>',
        };
      }

      const products = await kysely
        .selectFrom('product')
        .select(['id', 'name'])
        .where('name', 'ilike', `%${data.text}%`)
        .execute();

      if (products.length === 0) {
        return {
          method: 'sendMessage',
          chat_id: data.chatId,
          parse_mode: 'HTML',
          text: '<i>No products found. Type another keyword...</i>',
          reply_markup: {
            inline_keyboard: [
              [
                {
                  text: '❌ Cancel',
                  callback_data: 'cancel',
                },
              ],
            ],
          },
        };
      }

      await TelegramRepository.setChat({
        chatId: data.chatId,
        command: COMMAND,
        step: 3,
      });

      return {
        method: 'sendMessage',
        chat_id: data.chatId,
        text: 'Choose a product:',
        reply_markup: {
          inline_keyboard: [
            ...products.map((product) => [
              {
                text: product.name,
                callback_data: product.id.toString(),
              },
            ]),
            [
              {
                text: '❌ Cancel',
                callback_data: 'cancel',
              },
            ],
          ],
        },
      };
    }

    // Step 3
    case 3: {
      if (!data.isCallbackQuery) {
        await TelegramRepository.deleteChat(data.chatId);
        return {
          method: 'sendMessage',
          chat_id: data.chatId,
          parse_mode: 'HTML',
          text: '<i>Invalid command</i>',
          reply_markup: DEFAULT_REPLY_MARKUP,
        };
      }

      if (req.callback_query?.data === 'cancel') {
        await TelegramRepository.deleteChat(data.chatId);
        return {
          method: 'editMessageText',
          message_id: data.messageId,
          chat_id: data.chatId,
          parse_mode: 'HTML',
          text: '<i>Cancelled</i>',
        };
      }

      const product = await kysely
        .selectFrom('product')
        .select(['id', 'name'])
        .where('product.id', '=', req.callback_query!.data)
        .executeTakeFirstOrThrow();

      // Is it already in the list?
      const isExists =
        (
          await kysely
            .selectFrom('watch_list')
            .select((eb) =>
              eb.fn.count<number>('id').as('count'),
            )
            .where('chat_id', '=', data.chatId.toString())
            .where('product_id', '=', product.id.toString())
            .executeTakeFirstOrThrow()
        ).count > 0;

      if (isExists) {
        await TelegramRepository.deleteChat(data.chatId);
        return {
          method: 'editMessageText',
          message_id: data.messageId,
          chat_id: data.chatId,
          parse_mode: 'HTML',
          text: `<i>❌ ${product.name.toUpperCase()} is already in the list</i>`,
        };
      }

      // Add to the list
      await kysely
        .insertInto('watch_list')
        .values({
          chat_id: data.chatId,
          product_id: product.id,
          created_at: new Date(),
        })
        .executeTakeFirstOrThrow();

      await TelegramRepository.deleteChat(data.chatId);
      return {
        method: 'editMessageText',
        message_id: data.messageId,
        chat_id: data.chatId,
        parse_mode: 'HTML',
        text: `<i>✅ ${product.name.toUpperCase()} added to the list</i>`,
      };
    }

    default: {
      return {
        method: 'sendMessage',
        chat_id: data.chatId,
        parse_mode: 'HTML',
        text: '<i>Unhandled step</i>',
      };
    }
  }
}
