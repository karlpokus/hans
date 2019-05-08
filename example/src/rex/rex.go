// grabs a random user from url and prints their id and name
package main

import (
	"net/http"
	"os"
	"encoding/json"
	"time"
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
	err = json.NewDecoder(resp.Body).Decode(&user)
	return user, err
}

func main() {
	rand.Seed(time.Now().UnixNano())
	url := "https://jsonplaceholder.typicode.com/users"
	for {
		user, err := request(url)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			continue
		}
		os.Stdout.WriteString(user.Name)
		time.Sleep(8 * time.Second)
	}
}
