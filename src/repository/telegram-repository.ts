import { kysely } from '../database';

export class TelegramRepository {
  public static async setChat(data: {
    chatId: number;
    command: string;
    step: number;
    data?: Record<string, any>;
  }) {
    const timestamp = new Date();

    const chat = await kysely
      .selectFrom('chat')
      .select(['id'])
      .where('id', '=', data.chatId.toString())
      .executeTakeFirst();

    if (chat) {
      return await kysely
        .updateTable('chat')
        .set({
          command: data.command,
          step: data.step,
          data: data.data,
          updated_at: timestamp,
        })
        .where('chat.id', '=', data.chatId.toString())
        .returning([
          'chat.id',
          'chat.command',
          'chat.step',
          'chat.data',
        ])
        .executeTakeFirstOrThrow();
    }

    return await kysely
      .insertInto('chat')
      .values({
        id: data.chatId,
        command: data.command,
        step: data.step,
        data: data.data,
        created_at: timestamp,
      })
      .returning([
        'chat.id',
        'chat.command',
        'chat.step',
        'chat.data',
      ])
      .executeTakeFirstOrThrow();
  }

  public static async getChat(chatId: number) {
    return await kysely
      .selectFrom('chat')
      .select([
        'chat.id',
        'chat.command',
        'chat.step',
        'chat.data',
      ])
      .where('chat.id', '=', chatId.toString())
      .executeTakeFirst();
  }

  public static async deleteChat(chatId: number) {
    await kysely
      .deleteFrom('chat')
      .where('chat.id', '=', chatId.toString())
      .executeTakeFirst();
  }
}
