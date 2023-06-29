package main

import "github.com/Peltoche/neurone/src/server"

func main() {
	cfg := server.NewDefaultConfig()

	server.Start(cfg)
}
