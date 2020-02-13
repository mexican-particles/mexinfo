package mexinfo

func formatSlackMessage(query string) (*Message, error) {
	message := &Message{
		ResponseType: "in_channel",
		Text:         query,
	}
	return message, nil
}
