// Command api-compat-check compares a purecdk8s package with its upstream API.
package main

import apicheck "github.com/Chriscbr/purecdk8s/tools/api-compat-check"

func main() {
	apicheck.Main()
}
