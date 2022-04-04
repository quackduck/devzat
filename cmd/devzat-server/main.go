package main

func main() {
	var server DevzatServer

	if err := server.Init(); err != nil {
		panic(err)
	}
}
