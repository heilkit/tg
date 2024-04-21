package tg

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

const (
	ReactionLike       = "👍"
	ReactionDislike    = "👎"
	ReactionHeart      = "❤"
	ReactionOK         = "👌"
	ReactionFire       = "🔥"
	ReactionPray       = "🙏"
	ReactionStrawberry = "🍓"
	ReactionClown      = "🤡"
)

// React sets emoji/custom reaction to the give message. They are like 72 emojis, tg.Reaction* are the preset ones.
func (b *Bot) React(to *Message, reaction string, isBig ...bool) error {
	params := map[string]string{
		"chat_id":    strconv.FormatInt(to.Chat.ID, 10),
		"message_id": strconv.Itoa(to.ID),
		"reaction":   fmt.Sprintf("[{\"type\": \"%s\", \"emoji\": \"%s\"}]", reactionType(reaction), reaction),
	}
	if len(isBig) > 0 && isBig[0] {
		params["is_big"] = "true"
	}

	_, err := b.Raw("setMessageReaction", params)
	if err != nil {
		return wrapError(err)
	}

	return nil
}

func reactionType(reaction string) string {
	if utf8.RuneCountInString(reaction) <= 4 {
		return "emoji"
	}
	return "custom_emoji_id"
}
