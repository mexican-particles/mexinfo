package mexinfo

func formatSlackMessage(query string) (*Message, error) {
	message := &Message{
		ResponseType: "in_channel",
		// とりあえず適当に
		Text: "https://www.pref.hiroshima.lg.jp/uploaded/attachment/240510.pdf " + query,
	}
	return message, nil
}
