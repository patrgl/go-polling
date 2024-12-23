package main

import (
	"fmt"
	"net/http"
	"os"

	"go-polling/polls"
	"go-polling/vote"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	const base_url string = "localhost:8080/"

	os.Setenv("baseurl", "localhost:8080")

	baseUrl := os.Getenv("baseurl")

	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	/*
		db.AutoMigrate(&models.Poll{})
		db.AutoMigrate(&models.PollOption{})
		db.AutoMigrate(&models.VotedIp{})
	*/
	mux := http.NewServeMux()

	//routes
	mux.HandleFunc("/create", polls.CreatePoll())
	mux.HandleFunc("POST /submit-new-poll", polls.SubmitNewPoll(db, base_url))
	mux.HandleFunc("POST /close-poll", polls.ClosePoll(db))
	mux.HandleFunc("POST /get-results", polls.GetResults(db))
	mux.HandleFunc("/admin/{admin_poll_link}", polls.AdminPoll(db, base_url))
	mux.HandleFunc("/results/{poll_link}", polls.ViewResults(db, base_url))

	mux.HandleFunc("/vote/{poll_link}", vote.VotePage(db))
	mux.HandleFunc("POST /submit-vote", vote.SubmitVote(db, base_url))

	fmt.Printf("Serving on %s", baseUrl)
	err = http.ListenAndServe(baseUrl, mux)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed to start server")
		return
	}
}
