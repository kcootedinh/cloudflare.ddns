package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"net/http"
	"time"
)

type IPAddress struct {
	IP string `json:"ip"`
}

func main() {
	// create a new scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		// handle error
		panic(err)
	}

	// add a job to the scheduler
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				resp, err := http.Get("https://api.ipify.org?format=json")
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				// decode the response
				var ip IPAddress
				if err := json.NewDecoder(resp.Body).Decode(&ip); err != nil {
					panic(err)
				}

				// do things
				fmt.Println(time.Now(), ip)
			},
		),
	)
	if err != nil {
		// handle error
		panic(err)
	}

	// each job has a unique id
	fmt.Println(j.ID())

	// start the scheduler
	s.Start()

	// block until you are ready to shut down
	select {
	case <-time.After(time.Minute):
	}

	// when you're done, shut it down
	err = s.Shutdown()
	if err != nil {
		// handle error
		panic(err)
	}
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
