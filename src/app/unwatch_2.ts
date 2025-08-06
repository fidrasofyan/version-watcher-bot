import { DEFAULT_REPLY_MARKUP } from '../constant';
import { kysely } from '../database';
import { TelegramRepository } from '../repository';
import type {
  TelegramRequest,
  TelegramResponse,
} from '../types';

const COMMAND = 'unwatch_';

export async function unwatch2(
  req: TelegramRequest,
): Promise<TelegramResponse> {
  let chat = await TelegramRepository.getChat(
    req.message!.chat.id,
  );
  if (!chat) {
    chat = await TelegramRepository.setChat({
      chatId: req.message!.chat.id,
      command: COMMAND,
      step: 1,
      data: {},
    });
  }

  switch (chat.step) {
    // Step 1
    case 1: {
      const productName = req
        .message!.text?.replace('/unwatch_', '')
        .replaceAll('_', '-');

      if (!productName) {
        return {
          method: 'sendMessage',
          chat_id: req.message!.chat.id,
          parse_mode: 'HTML',
          text: '<i>Unknown command</i>',
          reply_markup: DEFAULT_REPLY_MARKUP,
        };
      }

      const product = await kysely
        .selectFrom('product')
        .select(['id', 'name'])
        .where('name', '=', productName)
        .executeTakeFirst();

      if (!product) {
        return {
          method: 'sendMessage',
          chat_id: req.message!.chat.id,
          parse_mode: 'HTML',
          text: '<i>Product not found</i>',
          reply_markup: DEFAULT_REPLY_MARKUP,
        };
      }

      await TelegramRepository.setChat({
        chatId: req.message!.chat.id,
        command: COMMAND,
        step: 2,
        data: {
          product,
        },
      });

      return {
        method: 'sendMessage',
        chat_id: req.message!.chat.id,
        parse_mode: 'HTML',
        text: `<i>Are you sure you want to remove ${product.name.toUpperCase()}?</i>`,
        reply_markup: {
          resize_keyboard: true,
          keyboard: [['Yes', 'No']],
        },
      };
    }

    // Step 2
    case 2: {
      if (req.message!.text !== 'Yes') {
        await TelegramRepository.deleteChat(
          req.message!.chat.id,
        );

        return {
          method: 'sendMessage',
          chat_id: req.message!.chat.id,
          parse_mode: 'HTML',
          text: '<i>Cancelled</i>',
          reply_markup: DEFAULT_REPLY_MARKUP,
        };
      }

      const chat = await TelegramRepository.getChat(
        req.message!.chat.id,
      );

      const result = await kysely
        .deleteFrom('watch_list')
        .where(
          'chat_id',
          '=',
          req.message!.chat.id.toString(),
        )
        .where(
          'product_id',
          '=',
          (chat!.data as any).product.id,
        )
        .executeTakeFirstOrThrow();

      await TelegramRepository.deleteChat(
        req.message!.chat.id,
      );

      if (result.numDeletedRows <= 0) {
        return {
          method: 'sendMessage',
          chat_id: req.message!.chat.id,
          parse_mode: 'HTML',
          text: '<i>Failed to remove product</i>',
          reply_markup: DEFAULT_REPLY_MARKUP,
        };
      }

      return {
        method: 'sendMessage',
        chat_id: req.message!.chat.id,
        parse_mode: 'HTML',
        text: `<i>${(chat!.data as any).product.name.toUpperCase()} removed from watch list</i>`,
        reply_markup: DEFAULT_REPLY_MARKUP,
      };
    }

    default: {
      return {
        method: 'sendMessage',
        chat_id: req.message!.chat.id,
        parse_mode: 'HTML',
        text: '<i>Unhandled step</i>',
        reply_markup: DEFAULT_REPLY_MARKUP,
      };
    }
  }
}
