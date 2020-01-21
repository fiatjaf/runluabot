package main

import (
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func handle(upd tgbotapi.Update) {
	if upd.Message != nil {
		handleMessage(upd.Message)
	} else if upd.InlineQuery != nil {
		handleInlineQuery(upd.InlineQuery)
	}
}

func handleMessage(message *tgbotapi.Message) {
	ret, err := runlua(message.Text)
	if err != nil {
		log.Warn().Err(err).Msg("message runlua")
		return
	}

	sendMessageAsReply(message.Chat.ID, ret, message.MessageID)
}

func handleInlineQuery(q *tgbotapi.InlineQuery) {
	code := q.Query

	ret, err := runlua(code)
	if err != nil {
		log.Warn().Err(err).Msg("inline runlua")
		return
	}

	bot.AnswerInlineQuery(tgbotapi.InlineConfig{
		InlineQueryID: q.ID,
		Results: []interface{}{
			tgbotapi.InlineQueryResultArticle{
				Type:                "article",
				ID:                  "result",
				Title:               ret,
				InputMessageContent: ret,
			},
			tgbotapi.InlineQueryResultArticle{
				Type:                "article",
				ID:                  "full",
				Title:               code + " => " + ret,
				InputMessageContent: code + " => " + ret,
			},
		},
		CacheTime: 180,
	})
}
