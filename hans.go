package main

import (
	"log"
	"os"
	"os/signal"
	"os/exec"
	"syscall"
	"time"
)

type hansLogger struct {
	log *log.Logger
}

func (h *hansLogger) Write(p []byte) (int, error) {
	h.log.Print(string(p))
	return len(p), nil
}

func main() {
	hans := &hansLogger{
		log: log.New(os.Stdout, "hans ", log.Ldate | log.Ltime),
	}

	pwd, _ := os.Getwd()
	app := pwd + "/apps/dummy"

	cmd := exec.Command(app, "foo", "5")
	cmd.Stdout = hans

	go func() {
		if err := cmd.Start(); err != nil {
			hans.log.Fatal(err)
		}
		if err := cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
			hans.log.Printf("err from cmd.Wait: %s", err.Error())
		}
		hans.log.Println("cmd.Wait done")
	}()

	// listen for SIGINT
	sigs := make(chan os.Signal, 1) // signals are strings
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs

		//hans.log.Println("the signal", sig)
		hans.log.Printf("attempting to kill %d", cmd.Process.Pid)

		if err := cmd.Process.Kill(); err != nil {
			hans.log.Printf("err from cmd.Process.Kill %s", err.Error())
		}

		time.Sleep(3 * time.Second)
		done <- true
	}()

	<-done
	hans.log.Println("hans exiting")
}
