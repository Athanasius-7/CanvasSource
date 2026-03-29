/*
Imran Qasimi
Canvas Source API
This module serves as the main writer for the retrieved
data from the Canvas API. Converting courses and assignments
to directories and md notes.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*
────────────────────────────────────────
Function creates a directory according to the string passed in.
MkdirAll() already has error handling in checking whether a given directory
exists or not.
────────────────────────────────────────
*/
func make_dir(path string) error {
	newpath := filepath.Join(".", path)
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Fatalf("There was error with creating the directories: %v", err)
	}
	return nil
}

/*
────────────────────────────────────────
Function creates and writes a todo markdown file
for all of the courses and their respective
assignments passed in.
────────────────────────────────────────
*/

func make_todo(courses *[]Course, path string) error {
	file, err := os.Create(path)
	// Error handling if the passed in path does not exist
	// or if there are issues with reading from the path.
	if err != nil {
		log.Fatalf("There was an error with creating a file for the	path passed in: %v\n", err)
		return err
	}
	// Iterate through the courses slice and create a title for each
	// course, then write the corresponding assignments for that course.
	var buffer strings.Builder
	for i := 0; i < len(*courses); i++ {
		// Make the title of the course center and give it a level 1 heading.
		var title string = fmt.Sprintf("<center><h1> %s </center></h1>\n\n", (*courses)[i].Name)
		buffer.WriteString(title)
		for j := 0; j < len((*courses)[i].Assignments); j++ {
			// Iterate through every assignment in this course slice and add it as a TODO
			// item to the buffer.
			buffer.WriteString(fmt.Sprintf("- [ ] [[%s/%s|%s]]\n", (*courses)[i].Name, (*courses)[i].Assignments[j].Title, (*courses)[i].Assignments[j].Title))
		}
	}
	defer file.Close()
	log.Printf("Successfully created and wrote to path: %v\n", file)
	// After generating all the text, write the file from the buffer
	_, err = file.WriteString(buffer.String())
	// Check if there was an error writing to the file.
	if err != nil {
		log.Fatalf("There was an error writing to the file. %v\n", err)
		return err
	}
	return err
}

/*
────────────────────────────────────────
Function generates a string that has an organized and formatted
markdown body for the specific parameter *Assignment that follows
the structure of the mock assignment in obsidian notes.
────────────────────────────────────────
*/

func assign_to_md(assignment *Assignment, path string) error {
	// Create the new file according to the path and assignment
	// passed in.
	var file_title string = fmt.Sprintf("%s.md", (*assignment).Title)
	file_title = strings.Replace(file_title, "/", "_", -1)
	newPath := filepath.Join(".", path, file_title)
	file, err := os.Create(newPath)
	// Check for errors.
	if err != nil {
		log.Fatalf("There was an error creating the file for the following"+
			" assignment, check integrity or filepath. %v\n%v\n", (*assignment).Title, err)
	}
	defer file.Close()
	var buffer strings.Builder
	buffer.WriteString("<center><h1> Description </center></h1>\n\n")
	buffer.WriteString((*assignment).Desc + "\n\n")
	buffer.WriteString(fmt.Sprintf("**Due: %s**", (*assignment).Due_date))
	_, err = file.WriteString(buffer.String())
	if err != nil {
		log.Fatalf("There was an error in writing to the file. %v\n", err)
	}
	return nil
}

/*
────────────────────────────────────────
Function creates the directories for each
course and then appends a markdown file for
each assignment within that specific course,
and to which directory.
────────────────────────────────────────
*/

func init_courses(courses *[]Course, path string) {
	// Check that the path does exist.
	fBool, err := FileExists(path)
	if !fBool || err != nil {
		log.Fatalf("There is in an error with the path passed. %v\n%v\n", path, err)
		return
	}
	for i := 0; i < len(*courses); i++ {
		err := make_dir(fmt.Sprintf("%s/%s", path, (*courses)[i].Name))
		if err != nil {
			return
		}
		for j := 0; j < len((*courses)[i].Assignments); j++ {
			assign_to_md(&(((*courses)[i]).Assignments[j]), filepath.Join(path, (*courses)[i].Name))
		}
	}
	return
}
