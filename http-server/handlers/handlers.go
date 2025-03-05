package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_final_project/api"
	http_server "go_final_project/http-server"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	FullOutput = iota
	DateSearch
	TextSearch
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TaskResponseID struct {
	ID string `json:"id"`
}

func ApiNextDate(w http.ResponseWriter, req *http.Request) {

	now := req.FormValue("now")
	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		log.Panic(err)
	}
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")
	answear, err := api.NextDate(nowTime, date, repeat)
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write([]byte(answear))
	if err != nil {
		log.Printf("error while writing response: %v", err)
	}

}

func PostTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http_server.ResponseJson("Error when reading from req.Body", http.StatusBadRequest, err, w)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		if task.Title == "" {
			http_server.ResponseJson("the Title field cannot be empty", http.StatusBadRequest, nil, w)
			return
		}
		var date string

		switch {
		case task.Date == time.Now().Format("20060102"):
			date = time.Now().Format("20060102")
		case task.Repeat == "" && task.Date == "":
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date == "":
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date != "":
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
				return
			}
			if dataTime.After(time.Now()) {
				date = task.Date
			} else {
				date, err = api.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http_server.ResponseJson("api.NextDate Error", http.StatusBadRequest, err, w)
					return
				}
			}
		case task.Repeat == "" && task.Date != "":
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
				return
			}
			if dataTime.Before(time.Now()) {
				date = time.Now().Format("20060102")
			} else {
				date = task.Date
			}
		default:
			date = task.Date
		}

		lastID, err := serverJob.AddTask(date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		var answear TaskResponseID = TaskResponseID{lastID}

		answearJSON, err := json.Marshal(&answear)
		if err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(answearJSON)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

func GetTasks(serverJob ServerJob, NumberOfOutTasksString string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		NumberOfOutTasksInt, err := strconv.Atoi(NumberOfOutTasksString)
		if err != nil {
			http_server.ResponseJson("NumberOfOutTasksString to NumberOfOutTasksInt convert error. Сheck the .env file.", http.StatusInternalServerError, err, w)
			return
		}
		reqQuery := req.URL.Query().Get("search")
		var code int
		_, err = time.Parse("02.01.2006", reqQuery)
		if reqQuery == "" {
			code = FullOutput
		} else if err == nil {
			code = DateSearch
		} else {
			code = TextSearch
		}

		var answearRows []Task
		if code == FullOutput {
			answearRows, err = serverJob.GetTasks(NumberOfOutTasksInt)
			if err != nil {
				http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
				return
			}
		} else {
			answearRows, err = serverJob.SearchTasks(code, reqQuery, NumberOfOutTasksInt)
			if err != nil {
				http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
				return
			}
		}

		if answearRows == nil {
			answearRows = make([]Task, 0)
		}
		var answearsStruck = map[string][]Task{"tasks": answearRows}
		jsonMsg, err := json.Marshal(answearsStruck)
		if err != nil {
			http_server.ResponseJson("error when using the Marshal function", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

func GetTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		qeryID := req.URL.Query().Get("id")
		if qeryID == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		answerRow, err := serverJob.GetTask(qeryID)
		if err != nil {
			http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
			return
		}
		jsonMsg, err := json.Marshal(answerRow)
		if err != nil {
			http_server.ResponseJson("Error when using the Marshal function", http.StatusInternalServerError, err, w)
			return

		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

func CorrectTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http_server.ResponseJson("Error when reading from req.Body", http.StatusBadRequest, err, w)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http_server.ResponseJson("deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		_, err = strconv.Atoi(task.ID) // проверка, что в поле ID передана цифра
		if err != nil {
			http_server.ResponseJson("wrong ID", http.StatusBadRequest, err, w)
			return

		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
			return
		}

		if task.Title == "" { // the Title value must always contain a description of the task. If it is empty, we return an error
			http_server.ResponseJson("The title is empty", http.StatusBadRequest, nil, w)
			return

		} // the date of the task must be greater than today. Otherwise we return an error
		if dateTime.Before(time.Now()) && dateTime.Equal(time.Now()) {
			http_server.ResponseJson("the date can't be less than today", http.StatusBadRequest, nil, w)
			return
		}

		if task.Repeat != "" {
			startChars := []rune{'y', 'd', 'm', 'w'}
			firstRuneRepeat := []rune(task.Repeat)[0]
			flagChek := true
			for _, s := range startChars {
				if s == firstRuneRepeat {
					flagChek = false
				}
			}
			if flagChek {
				http_server.ResponseJson("the rule for repetition has the wrong format.", http.StatusBadRequest, nil, w)
				return
			}
		}

		err = serverJob.UpdateTask(task)
		if err != nil {
			http_server.ResponseJson("error when updating data on the server", http.StatusInternalServerError, err, w)
			return
		}
		jsonMsg, err := json.Marshal(Task{})
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}

	}
}

func DoneTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		_, err := strconv.Atoi(idDone) // проверка, на то что в поле ID передано число
		if err != nil {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, err, w)
			return
		}
		task, err := serverJob.GetTask(idDone) // проверка, что задание существует в базе данных
		if task.ID == "" {
			http_server.ResponseJson("id not found in the database", http.StatusBadRequest, err, w)
			return
		}
		if task.Repeat == "" {
			err = serverJob.DeleteTask(idDone)
			if err != nil {
				http_server.ResponseJson("failed to delete the task", http.StatusInternalServerError, err, w)
				return
			}
		} else {
			newDateString, err := api.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http_server.ResponseJson("api.NextDate Error", http.StatusInternalServerError, err, w)
				return
			}
			task.Date = newDateString
			err = serverJob.UpdateTask(task)
			if err != nil {
				http_server.ResponseJson("data update error", http.StatusInternalServerError, err, w)
				return
			}
		}
		var answear = map[string]any{}
		jsonMsg, err := json.Marshal(answear)
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

func DeleteTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		_, err := strconv.Atoi(idDone) // проверка, передано ли в поле ID цифра
		if err != nil {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, err, w)
			return
		}

		err = serverJob.DeleteTask(idDone)
		if err != nil {
			http_server.ResponseJson("failed to delete the task", http.StatusBadRequest, err, w)
			return
		}

		var answear = map[string]any{}
		jsonMsg, err := json.Marshal(answear)
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}
