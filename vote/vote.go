package vote

import (
	"fmt"
	"go-polling/ips"
	"go-polling/models"
	"net/http"

	"gorm.io/gorm"
)

func VotePage(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_link := r.PathValue("poll_link")

		query_poll := models.Poll{}
		res := db.First(&query_poll, "poll_link = ?", poll_link)
		if res.Error != nil {
			fmt.Fprint(w, `<h1>Invalid Poll Link!</h1>`)
			return
		}

		if !query_poll.Open {
			fmt.Fprint(w, `<h1>Poll closed!</h1>`)
			return
		}

		if query_poll.LimitIps != "none" {
			voter_ip := ips.GetIp(r)
			query_ip := models.VotedIp{}
			res = db.First(&query_ip, "ip = ? AND poll_id = ?", voter_ip, query_poll.ID)
			if res.Error == nil {
				fmt.Fprint(w, `<h1>You have already voted on this poll!</h1>`)
				return
			}
		}

		query_options := []models.PollOption{}
		db.Where("poll_id = ?", query_poll.ID).Find(&query_options)

		results := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Vote</title>
			<script src="https://unpkg.com/htmx.org@2.0.1"></script>
			<style>
				body{
					height: 100vh;
					display: flex;
					align-items: center;
					justify-content: center;
					font-family: 'Railway', sans-serif;
					text-align: center;
				}

				.vote-holder {
					text-align: left;
				}
				
				button {
					margin: 5px;
					font-size: 20px;
				}
			</style>
		</head>
		<body>
		<div class="poll-holder" id="poll-holder"><h1>%s</h1>`, query_poll.Name)

		for _, option := range query_options {
			to_insert := fmt.Sprintf(`
			<form hx-post="/submit-vote" hx-target="#poll-holder" hx-swap="innerHTML">
			<input type="hidden" name="option-id" value="%d">
			<input type="hidden" name="poll-id" value="%d">
			<button type="submit">%s</button></form>
			`, option.ID, query_poll.ID, option.Name)
			results = results + to_insert
		}
		results = results + `</div></body></html>`
		fmt.Fprint(w, results)
	}
}

func SubmitVote(db *gorm.DB, base_url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_id := r.PostFormValue("poll-id")
		option_id := r.PostFormValue("option-id")

		query_poll := models.Poll{}
		poll_result := db.First(&query_poll, "id = ?", poll_id)
		if poll_result.Error != nil {
			fmt.Println(w, "Invalid poll ID: %v", poll_id)
		}

		query_option := models.PollOption{}
		option_result := db.First(&query_option, "id = ?", option_id)
		if option_result.Error != nil {
			fmt.Println(w, "Invalid option ID: %v", option_id)
		}

		query_option.Votes += 1
		db.Save(&query_option)

		if query_poll.LimitIps != "none" {
			add_ip := models.VotedIp{
				Ip:     ips.GetIp(r),
				PollId: int(query_poll.ID),
			}
			db.Create(&add_ip)
		}
		complete_url := base_url + "results/" + query_poll.PollLink
		fmt.Fprintf(w,
			`<h1>Vote Submitted!</h1>
			View results at: <a href="%v">%v</a>`, complete_url, complete_url)
	}
}
