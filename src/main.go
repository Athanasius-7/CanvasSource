/*
Imran Qasimi
Canvas Source Main
*/
package main

import (
	"database/sql"
	"fmt"
	_"github.com/mattn/go-sqlite3"
	"time"
	"log"
)

func main() {
	fmt.Printf("The user's OS and Linux Distro is: %s\n", GetDistro())
	begin := time.Now()
	cookie := GetSessionCookie()
	var courses []Course
	err := GetCourses(&courses, cookie)
	if err != nil {
		log.Fatalln(err)
	}
	make_todo(&courses, "assignments/Assignment_List.md")
	fmt.Printf("The time it took to create courses and make the todo list was: %v\n", time.Since(begin))
	init_courses(&courses, "assignments")
	fmt.Printf("The time it took to create all assignments was %v:\n", time.Since(begin))
	// Sqlite3 Code
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		log.Fatalf("There was an error with open the database.\n%v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("There was an error ping the database.\n%v", err)
	}
	fmt.Printf("Successfully pinged the database.\n")
	return
}
