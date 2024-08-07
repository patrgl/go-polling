package polls

import (
	"fmt"
	"go-polling/models"
	"math/rand"
	"net/http"
	"strings"
	"text/template"

	"gorm.io/gorm"
)

func CreateLink(admin_or_poll string, db *gorm.DB) string {
	for {
		//create short
		const chars string = "abcdefghijklmnopqrstuvwxyz0123456789"
		short := make([]byte, 6)
		for i := 0; i < 6; i++ {
			short[i] += chars[rand.Intn(len(chars))]
		}
		link := string(short)

		query_poll := models.Poll{}
		field_to_query := ""
		if admin_or_poll == "poll" {
			field_to_query = "poll_link = ?"
		}
		if admin_or_poll == "admin" {
			field_to_query = "admin_link = ?"
		}

		result := db.First(&query_poll, field_to_query, link)
		if result.Error != nil {
			return link
		}
		fmt.Println("Randomly generated link already exists, trying again...")

	}

}

func CreatePoll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/create.html"))
		tmpl.Execute(w, nil)
	}
}

func AdminPoll(db *gorm.DB, base_url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin_link := r.PathValue("admin_poll_link")

		query_poll := models.Poll{}
		pollResult := db.First(&query_poll, "admin_link = ?", admin_link)
		if pollResult.Error != nil {
			fmt.Fprint(w, `<h1>Invalid Poll Link!</h1>`)
		}

		//package additional data
		query_poll.CompletePollLink = base_url + "vote/" + query_poll.PollLink

		tmpl := template.Must(template.ParseFiles("templates/admin.html"))
		tmpl.Execute(w, query_poll)
	}
}

func SubmitNewPoll(db *gorm.DB, base_url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_name := r.PostFormValue("poll-name")
		fmt.Println(poll_name)
		ip_limit := r.PostFormValue("ip-limit")
		poll_options := r.PostFormValue("poll-options")

		submit_poll := models.Poll{
			Name:      poll_name,
			Open:      true,
			LimitIps:  ip_limit,
			PollLink:  CreateLink("poll", db),
			AdminLink: CreateLink("admin", db),
		}
		s := db.Create(&submit_poll)
		if s.Error != nil {
			fmt.Println("Error submitting new poll")
		}

		for _, option := range strings.Split(poll_options, ",") {
			submit_option := models.PollOption{
				PollId: int(submit_poll.ID),
				Name:   option,
				Votes:  0,
			}
			ores := db.Create(&submit_option)
			if ores.Error != nil {
				fmt.Printf("Error submitting option %s", option)
			}
		}

		complete_admin_link := base_url + "admin/" + submit_poll.AdminLink
		complete_vote_link := base_url + "vote/" + submit_poll.PollLink

		fmt.Fprintf(w, `Public Poll Link: <a href="%s">%s</a><br>Private Poll Admin Link: <a href="%s">%s</a>`,
			complete_vote_link,
			complete_vote_link,
			complete_admin_link,
			complete_admin_link)
	}
}

func ClosePoll(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_id := r.PostFormValue("poll-id")

		query_poll := models.Poll{}
		result := db.First(&query_poll, "id = ?", poll_id)
		if result != nil {
			fmt.Fprint(w, `<h1>Unable to close poll (invalid poll ID)</h1>`)
		}

		query_poll.Open = false
		db.Save(&query_poll)
		fmt.Fprint(w, `<b>Poll Closed</b>`)
	}
}

func GetResults(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_id := r.PostFormValue("poll-id")

		query_options := []models.PollOption{}
		db.Where("poll_id = ?", poll_id).Find(&query_options)

		results := ``

		for _, option := range query_options {
			to_insert := fmt.Sprintf(`<b>%s:</b> %d<br>`, option.Name, option.Votes) //s for string, d for integers
			results += to_insert
		}
		fmt.Fprint(w, results)
	}
}

func ViewResults(db *gorm.DB, base_url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		poll_link := r.PathValue("poll_link")

		query_poll := models.Poll{}
		pollResult := db.First(&query_poll, "poll_link = ?", poll_link)
		if pollResult.Error != nil {
			fmt.Fprint(w, `<h1>Invalid Poll Link!</h1>`)
		}

		query_poll.CompletePollLink = base_url + "vote/" + query_poll.PollLink

		tmpl := template.Must(template.ParseFiles("templates/results.html"))
		tmpl.Execute(w, query_poll)
	}
}
