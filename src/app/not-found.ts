import { DEFAULT_REPLY_MARKUP } from "../constant";
import type { TelegramRequest, TelegramResponse } from "../types";

export function notFound(req: TelegramRequest): TelegramResponse {
  if (req.message) {
    return {
      method: "sendMessage",
      chat_id: req.message.chat.id,
      parse_mode: "HTML",
      text: "<i>Unknown command</i>",
      reply_markup: DEFAULT_REPLY_MARKUP,
    };
  } else if (req.callback_query) {
    return {
      method: "editMessageText",
      message_id: req.callback_query.message!.message_id,
      chat_id: req.callback_query.from.id,
      parse_mode: "HTML",
      text: "<i>Invalid session</i>",
    };
  }

  throw new Error("Unknown request type");
}
