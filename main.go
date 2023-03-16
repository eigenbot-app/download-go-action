package main

import (
	"flag"
)

var (
	owner, repo, name, token string
)

func main() {
	flag.StringVar(&owner, "owner", "", "")
	flag.StringVar(&repo, "repo", "", "")
	flag.StringVar(&name, "name", "", "")
	flag.StringVar(&token, "token", "", "")
	flag.Parse()
}
