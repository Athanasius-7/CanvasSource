/*
Imran Qasimi
Canvas Source File Reader
This module contains the file reading functionality for
the project, including retrieving cookies, rows, and
system info checking.
*/
package main

import (
	jsonv2 "encoding/json/v2"
	"errors"
	_ "fmt"
	"github.com/giulianopz/go-dejsonlz4/jsonlz4"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
)

type Cookies struct {
	Cookies []Cookie `json:"cookies"`
}

type Cookie struct {
	Host  string `json:"host"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

/*
────────────────────────────────────────
Authentication function to make sure
a path exists in the user's system.
────────────────────────────────────────
*/

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("File does not exist. Ensure correct path has been entered. Path:%v\n%v\n", path, err)
			return false, nil
		}
		log.Fatalln("Check the validity of the file you are reading from.")
		return false, err
	}
	return true, nil
}

/*
────────────────────────────────────────
Returns the current default FireFox Profile
path as a string.
────────────────────────────────────────
*/

func FireFoxPath() string {
	switch GetDistro() {
	case "NAME=\"Arch Linux\"":
		user, _ := user.Current()
		ini_path := filepath.Join(user.HomeDir, ".config/mozilla/firefox/profiles.ini")
		cfg, err := ini.Load(ini_path)
		if err != nil {
			log.Fatalf("There was an error reading the profiles for FireFox:%v\n", err)
		}
		default_profile := cfg.Section("Profile0").Key("Path").String()
		default_path := filepath.Join(user.HomeDir, ".config/mozilla/firefox/", default_profile, "sessionstore-backups/recovery.jsonlz4")
		return default_path
	case "NAME=\"Ubuntu\"":
		user, _ := user.Current()
		ini_path := filepath.Join(user.HomeDir, "snap/firefox/common/.mozilla/firefox/profiles.ini")
		cfg, err := ini.Load(ini_path)
		if err != nil {
			log.Fatalf("There was an error reading the profiles for FireFox:%v\n", err)
		}
		default_profile := cfg.Section("Profile0").Key("Path").String()
		default_path := filepath.Join(user.HomeDir, "snap/firefox/common/.mozilla/firefox/", default_profile, "sessionstore-backups/recovery.jsonlz4")
		return default_path
	default:
		return "No SystemD OS detected."
	}
}

/*
────────────────────────────────────────
Function reads the active recovery.jsonlz4
when successfully read, it return a a struct of Cookies
that contains all the active cookies in the users
browser session.
────────────────────────────────────────
*/

func SessionCookies(path string) (*Cookies, error) {
	file_exists, err := FileExists(path)
	if !file_exists {
		return nil, err
	}
	session_lz4, _ := os.ReadFile(path)
	content, err := jsonlz4.Uncompress(session_lz4)
	if err != nil {
		log.Fatalln("Error: File was not decoded properly.")
		return nil, err
	}
	var kookies Cookies
	err = jsonv2.Unmarshal(content, &kookies)
	if err != nil {
		log.Fatalln("Error: File was not unmarshaled properly.")
		return nil, err
	}
	return &kookies, err
}

/*
────────────────────────────────────────
Function returns the user's linux distro.
Parses hostnamectl cli response and maps
it onto a map of [string]string that then
returns the Operating System response value.
────────────────────────────────────────
*/
// TODO: Use cat /etc/os-release for cli command, then parse the output
// using regex.
func GetDistro() string {
	// cat the os-release file in order to get the user's system info,
	// specifically for linux machines.
	cmdStruct := exec.Command("/bin/bash", "-c", "cat /etc/os-release")
	output_0, err := cmdStruct.Output()
	if err != nil {
		log.Fatalf("There was an error with executing cat /etc/os-release.\n%v", err)
	}
	var os_release string = string(output_0)
	regex, _ := regexp.Compile("^NAME=\"(.*?)\"")
	return regex.FindString(string(os_release))
}
