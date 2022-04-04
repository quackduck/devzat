package interfaces

type Room interface {
	hasName
	hasBot
	hasServer
	hasAdmins
	managesUsers
	hasSlackIntegration
	checksName
	managesColors

	Broadcast(senderName, msg string)
	BroadcastNoSlack(senderName, msg string)
	Cleanup()
	Join(User)

	ParseUserInput(line string, u User) error
	PrintUsers() string
}
