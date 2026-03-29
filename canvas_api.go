/*
Imran Qasimi
Canvas Source API
This module serves as the primary info retrieval for
the user's canvas data, primarily through API calls.
*/

package main

import (
	jsonv2 "encoding/json/v2"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
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
		log.Fatalln("GetSessionCookie() --> Did not recieve correct value.")
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
	// make request.
	resp, err := http.Get(url)
	// check if we had an error.
	if err != nil {
		log.Fatalln(err)
	}
	text_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var user User
	jsonv2.Unmarshal([]byte(string(text_body)), &user)
	return user
}

func GetCourses(courses *[]Course, cookie *http.Cookie) error {
	var url string = fmt.Sprintf("https://mpc.instructure.com/api/v1/courses/?per_page=100")
	req, err := GetRequest(cookie, "GET", url)
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error in retrieving the courses.\n")
		log.Fatalf("The canvas session may be inactive within your browser. Make sure you are logged into your school's canvas domain: ERR: %v\n", err)
		return nil
	}
	text_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var buffer []Course
	begin := time.Now()
	err = jsonv2.Unmarshal([]byte(string(text_body)), &buffer)
	fmt.Printf("Time to unmarshal JSON into []buffer.%v\n", time.Since(begin))
	if err != nil {
		log.Fatalln(err)
	}
	var wg sync.WaitGroup
	begin = time.Now()
	// check whether a specific course is restricted, if not,
	// then do not add to our final course list.
	for i := 0; i < len(buffer); i++ {
		if !buffer[i].Restricted {
			*courses = append(*courses, buffer[i])
		}
	}
	for i := 0; i < len(*courses); i++ {
		wg.Add(1)
		go GetCourseAssignments((&(*courses)[i]), &((*courses)[i].Assignments), cookie, &wg)
	}
	wg.Wait()
	fmt.Printf("Time to unmarshal all asssignments into []courses.%v\n", time.Since(begin))
	return err
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
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error in retrieving assignments for [course_id: %d]\tStatus Code:%d\n", course.Course_ID, resp.StatusCode)
		log.Fatalf("The canvas session may be inactive within your browser. Make sure you are logged into your school's canvas domain: ERR: %v\n", err)
		return nil
	}
	text_body, err := io.ReadAll(resp.Body)
	// check if we had an error.
	if err != nil {
		log.Fatalln(err)
	}
	err = jsonv2.Unmarshal([]byte(string(text_body)), &assignments)
	if err != nil {
		log.Fatalln(err)
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
