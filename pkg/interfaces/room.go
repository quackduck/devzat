package interfaces

type Room interface {
	hasName
	hasBot
	hasServer
	hasAdmins
	managesUsers

	Broadcast(senderName, msg string)
	BroadcastNoSlack(senderName, msg string)
	Cleanup()

	ParseUserInput(line string, u User) error
	PrintUsersInRoom() string
}
