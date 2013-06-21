package chatserver

import (
	"log"
	"strconv"
	"strings"
)

func init() {
	Cmds := [2]string{"help", "info"}
	log.Println(Cmds)
}

func checkCmd(text string) bool {
	if text[0] == '/' {
		return true
	} else {
		return false
	}
}

func (u *OnlineUser) cmdResult(text string) string {
	text = strings.Trim(text, "/")
	var result string

	switch text {
	case "help":
		result = Help()
	case "info":
		result = u.Room.Info()
	default:
		result = NoCmd()
	}

	return result
}

func NoCmd() string {
	var text = "\n Not found you command \n please you /help get help"
	return text
}

func Help() string {
	text := `
    /help    get help
    /info    get room info    
    `
	return text
}

func (r *ActiveRoom) Info() string {
	var text string
	status, logStatus := "ON", "ON"

	if !r.Status {
		status = "OFF"
	}

	if !r.SaveLogs {
		logStatus = "OFF"
	}

	text = "\n Room is: " + r.RoomKey
	text = text + "\n Room user count is:" + strconv.Itoa(r.Users.Len())
	text = text + "\n Room status is:" + status
	text = text + "\n Room log is:" + logStatus

	return text
}
