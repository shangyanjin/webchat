package chatserver

import (
	"code.google.com/p/go.net/websocket"
	"webchat/app/model"
	//"strconv"
	"log"
	"time"
)

type OnlineUser struct {
	Id         int
	Connection *websocket.Conn
	Send       chan *Event
	Room       *ActiveRoom
	Info       *UserInfo
}

type UserInfo struct {
	Name   string
	Email  string
	Avatar string
}

func NewOnlineUser(user *model.User, ws *websocket.Conn, room *ActiveRoom) *OnlineUser {
	onlineUser := &OnlineUser{
		Id:         user.Id,
		Connection: ws,
		Send:       make(chan *Event, 512),
		Room:       room,

		Info: &UserInfo{
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.AvatarUrl(),
		},
	}

	return onlineUser
}

func (u *OnlineUser) PushToClient() {
	for b := range u.Send {
		err := websocket.JSON.Send(u.Connection, b)
		if err != nil {
			break
		}
	}
}

func (u *OnlineUser) PullFromClient() {
	for {
		var event Event
		if err := websocket.JSON.Receive(u.Connection, &event); err != nil {
			log.Println("Receive occur some error", err.Error())
			return
		}

		event.Created = time.Now()
		event.User = u.Info
		u.HandleMessage(&event)
	}
}

func (u *OnlineUser) HandleMessage(event *Event) {
	if checkCmd(event.Text) {
		u.ProcessCmd(event.Text)
	} else {
		if u.Room.Status {
			u.Room.Broadcast <- event
		}
		if u.Room.SaveLogs && u.Room.Status {
			u.SaveMessageToRedis(event)
		}
	}
}

func (u *OnlineUser) ProcessCmd(Text string) {
	resultText := u.cmdResult(Text)

	var event *Event = &Event{
		Type:    "cmd",
		Text:    resultText,
		Created: time.Now(),
		User:    u.Info,
	}

	u.Send <- event
}

func (u *OnlineUser) SaveMessageToRedis(event *Event) {
	// save to redis list
	// format: "text|lds|asd"
	listKey := "room:" + u.Room.RoomKey
	//time := event.Created.Unix()
	content := event.Type + "|" + event.User.Name + "|" + event.Text + "|" + event.Created.String()
	redisClient.Lpush(listKey, []byte(content))
}

func (u *OnlineUser) Close() {
	// clear resource when user conn close
	// close conn
	if err := u.Connection.Close(); err != nil {
		log.Println("close conn faild")
	}

	// close channel
	close(u.Send)

	// send levae message to other client
	event := &Event{
		Type:    "leave",
		Text:    u.Info.Name + " has leave room",
		User:    u.Info,
		Created: time.Now(),
	}

	u.Room.Broadcast <- event
}
