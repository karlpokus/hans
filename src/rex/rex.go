// grabs a random user from url and prints their id and name
package main

import (
	"net/http"
	"os"
	"encoding/json"
	"time"
	"log"
	"math/rand"
	"strconv"
)

type User struct {
	Name string
	Id int
}

func randInt(min, max int) int {
	return min + rand.Intn(max - min)
}

func randUser(url string) string {
	return url + "/" + strconv.Itoa(randInt(1, 10))
}

func request(url string) (User, error) {
	var user User
	resp, err := http.Get(randUser(url))
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return user, err
	}
	return user, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)
	url := "https://jsonplaceholder.typicode.com/users"
	for {
		user, err := request(url)
		if err != nil {
			stderr.Println(err)
		} else {
			stdout.Println(user.Id, user.Name)
		}
		time.Sleep(5 * time.Second)
	}
}
