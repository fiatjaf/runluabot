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

	sendMessageAsReply(message.Chat.ID, `<pre><code class="language-json">`+ret+`</code></pre>`, message.MessageID)
}

func handleInlineQuery(q *tgbotapi.InlineQuery) {
	code := q.Query

	ret, err := runlua(code)
	if err != nil {
		log.Warn().Err(err).Msg("inline runlua")
		return
	}

	_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
		InlineQueryID: q.ID,
		Results: []interface{}{
			tgbotapi.NewInlineQueryResultArticleHTML("result", ret,
				`<pre><code class="language-json">`+ret+`</code></pre>`,
			),
			tgbotapi.NewInlineQueryResultArticleHTML("full", code+" → "+ret,
				`<code>`+code+`</code> => <code class="language-json">`+ret+`</code>`,
			),
		},
		IsPersonal: false,
		CacheTime:  30,
	})
	if err != nil {
		log.Warn().Err(err).Msg("inline results")
	}
}
