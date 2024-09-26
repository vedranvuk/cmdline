package cmdutils

type (
	Topic = string

	HelpTopic struct {
		Name    string
		Title   string
		Content string
	}

	Help map[Topic]HelpTopic
)
