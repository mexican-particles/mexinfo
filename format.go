package mexinfo

func formatSlackMessage(link string) (*Message, error) {
	message := &Message{
		ResponseType: "in_channel",
		Text:         link,
		UnfurlLinks:  true,
	}
	return message, nil
}
