/*
Imran Qasimi
Canvas Source Main
*/
package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	// "runtime/trace"
	"time"
)

func main() {
	// Temp Code for Future Traces.
	/* file, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("There was an error with creating the trace.out file: %v\n", err)
	}
	defer file.Close()
	err = trace.Start(file)
	if err != nil {
		log.Fatalf("There was an error with starting the trace: %v\n", err)
	}
	defer trace.Stop() */
	db, err := initDB("./cache.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = initSchema(db); err != nil {
		log.Fatalf("There was an error with initSchema.%v\n", err)
	}
	user, _ := user.Current()
	homeDir := user.HomeDir
	assign_path := filepath.Join(homeDir, "assignments")
	err = os.MkdirAll(assign_path, os.ModePerm)
	if err != nil {
		log.Fatalf("There was an error with creating  the assignments path.%v\n", err)
	}
	initCache(db)
	cache := &Cache{}
	err = GetCache(cache, db, 1)
	if err != nil {
		log.Fatal(err)
	}
	begin := time.Now()
	cookie := GetSessionCookie()
	var courses []Course
	err = GetCourses(&courses, cookie)
	if err != nil {
		log.Fatalln(err)
	}
	todo_path := filepath.Join(assign_path, "TODO.md")
	make_todo(&courses, todo_path)
	// fmt.Printf("The time it took to create courses and make the todo list was: %v\n", time.Since(begin))
	init_courses(&courses, assign_path)
	fmt.Printf("The entire program took: %v\n", time.Since(begin))
	// Sqlite3 Code
	return
}
