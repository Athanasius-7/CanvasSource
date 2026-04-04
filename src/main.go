/*
Imran Qasimi
Canvas Source Main
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	db, err := initDB("./cache.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = initSchema(db); err != nil {
		log.Fatalf("There was an error with initSchema.%v\n", err)
	}
	initCache(db)
	cache := &Cache{}
	err = GetCache(cache, db, 1)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(cache.folder_path, os.ModePerm)
	if err != nil {
		log.Fatalf("There was an error with creating  the assignments path.%v\n", err)
	}
	begin := time.Now()
	var courses []Course
	status, err := GetCourses(&courses, &cache.cookie)
	if err != nil || status == UNSUCCESS {
		log.Fatal(err)
	}
	switch status{
		case SUCCESS:
		log.Printf("Course retrieval was successful -> return status: %d", status)
	case UPDATEREQ:
		log.Printf("Course retrieval was successful -> updating cache: %d", status)
		initCache(db)
	}
	todo_path := filepath.Join(cache.folder_path, "TODO.md")
	make_todo(&courses, todo_path)
	// fmt.Printf("The time it took to create courses and make the todo list was: %v\n", time.Since(begin))
	init_courses(&courses, cache.folder_path)
	fmt.Printf("The entire program took: %v\n", time.Since(begin))
	// Sqlite3 Code
	return
}
