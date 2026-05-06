package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("sweet24412"), bcrypt.DefaultCost)
	fmt.Print(string(hash))
}
