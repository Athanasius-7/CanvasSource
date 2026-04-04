/*
Imran Qasimi
Canvas Source API
This module serves as the primary info retrieval for
the user's canvas data, primarily through API calls.
*/

package main

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

type FunctionStatus int 
const (
	SUCCESS = iota
	UPDATEREQ
	UNSUCCESS
)
/*
────────────────────────────────────────
Struct Defintions for Canvas Types
────────────────────────────────────────
*/

type Assignment struct {
	Title    string `json:"name"`
	Due_date string `json:"due_at"`
	Desc     string `json:"description"`
	// Points    float64 `json:"points_possible"`
	// Grade     float64 `json:"grader_count"`
	// Graded    bool    `json:"graded_submissions_exist"`
	// Completed bool    `json:"has_submitted_submissions"`
}

type User struct {
	User_ID int `json:"id"`
}
type Course struct {
	Course_ID   int    `json:"id"`
	Name        string `json:"name"`
	State       string `json:"workflow_state"`
	Restricted  bool   `json:"access_restricted_by_date"`
	Assignments []Assignment
}

/*
────────────────────────────────────────
API Functions to Retrieve Data.
────────────────────────────────────────
*/

func GetSessionCookie() *http.Cookie {
	var session_path string = FireFoxPath()
	kookies, err := SessionCookies(session_path)
	if err != nil {
		log.Fatal("GetSessionCookie() --> Did not recieve correct value.")
		return nil
	}
	var target Cookie
	for i := 0; i < len(kookies.Cookies); i++ {
		if (kookies.Cookies)[i].Name == "canvas_session" && (kookies.Cookies)[i].Host == "mpc.instructure.com" {
			target = (kookies.Cookies)[i]
		}
	}
	var session_cookie = &http.Cookie{
		Name:  target.Name,
		Value: target.Value,
	}
	return session_cookie
}

func GetRequest(cookie *http.Cookie, method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("There was an error with generating the NewRequest.%v\n", err)
		return nil, err
	}
	req.AddCookie(cookie)
	return req, err
}

func GetUser() User {
	var url string = "https://learn.canvas.net/api/v1/users/self"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	text_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	sonic.Unmarshal([]byte(string(text_body)), &user)
	return user
}

func GetCourses(courses *[]Course, cookie *http.Cookie) (int, error) {
	var url string = fmt.Sprintf("https://mpc.instructure.com/api/v1/courses/?per_page=100")
	req, err := GetRequest(cookie, "GET", url)
	if err != nil {
		return UNSUCCESS, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return UNSUCCESS, err
	}
	defer resp.Body.Close()
	var status int = SUCCESS
	// The cookie passed in is most likely expired -> Update
	if resp.StatusCode != 200 {
		log.Print("Updating Cookie.\n")
		cookie = GetSessionCookie()
		req, err := GetRequest(cookie, "GET", url)
		if err != nil {
			return UNSUCCESS, err
		}
		resp, err = client.Do(req)
		if err != nil {
			return UNSUCCESS, err
			}
		// User's session is not active, therefore kill program
		if resp.StatusCode != 200 {
			log.Fatal("Ensure your canvas session is active.\n")
			return UNSUCCESS, nil
		}
		// Update flag -> Update Cache in main
		status = UPDATEREQ
	}
	defer resp.Body.Close()
	text_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UNSUCCESS, err
	}
	var buffer []Course
	err = sonic.Unmarshal([]byte(string(text_body)), &buffer)
	if err != nil {
		return UNSUCCESS, err
	}
	var wg sync.WaitGroup
	// check whether a specific course is restricted, if not,
	// then do not add to our final course list.
	// TODO: Add filter for only current term courses.
	for i := 0; i < len(buffer); i++ {
		if !buffer[i].Restricted {
			*courses = append(*courses, buffer[i])
		}
	}
	// add a fixed of groups to wg.
	wg.Add(len(*courses))
	bar := progressbar.Default(int64(len((*courses))))
	for i := 0; i < len(*courses); i++ {
		go GetCourseAssignments((&(*courses)[i]), &((*courses)[i].Assignments), cookie, &wg)
		bar.Add(1)
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait()
	return status, nil
}

/*
────────────────────────────────────────
Function populates the Assignments field of a passed in
course, returns an error if there was a mistake, otherwise nil
────────────────────────────────────────
*/

func GetCourseAssignments(course *Course, assignments *[]Assignment, cookie *http.Cookie, wg *sync.WaitGroup) error {
	defer wg.Done()
	var url string = fmt.Sprintf("https://mpc.instructure.com/api/v1/courses/%d/assignments", course.Course_ID)
	req, err := GetRequest(cookie, "GET", url)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// The cookie passed in is most likely expired -> Update
	if resp.StatusCode != 200 {
		log.Print("Updating Cookie.\n")
		cookie = GetSessionCookie()
		req, err := GetRequest(cookie, "GET", url)
		if err != nil {
			return err
		}
		resp, err = client.Do(req)
		if err != nil {
			return err
			}
		// User's session is not active, therefore kill program
		if resp.StatusCode != 200 {
			log.Fatal("Ensure your canvas session is active.\n")
			return nil
		}
	}
	text_body, err := io.ReadAll(resp.Body)
	// check if we had an error.
	if err != nil {
		log.Fatal(err)
	}
	err = sonic.Unmarshal(text_body, &assignments)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

/*
────────────────────────────────────────
Helper Functions to Print Type Fields
────────────────────────────────────────
*/

func course_to_str(course *Course) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("───── Course Fields: ─────\n"))
	struct_interface := reflect.TypeOf(*course)
	struct_values := reflect.ValueOf(*course)
	for i := 0; i < struct_interface.NumField(); i++ {
		field := struct_interface.Field(i)
		value := struct_values.Field(i)
		sb.WriteString(fmt.Sprintf("%v: %v\n", field.Name, value))
	}
	return sb.String()
}

func assignment_to_str(response *Assignment) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("───── Assignment Fields: ─────\n"))
	struct_interface := reflect.TypeOf(*response)
	struct_values := reflect.ValueOf(*response)
	for i := 0; i < struct_interface.NumField(); i++ {
		field := struct_interface.Field(i)
		value := struct_values.Field(i)
		sb.WriteString(fmt.Sprintf("%v: %v\n", field.Name, value))
	}
	return sb.String()
}
