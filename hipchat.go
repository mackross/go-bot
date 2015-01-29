package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mackross/go-bot/chat"

	"github.com/mackross/go-hipchat"
	api "github.com/tbruyelle/hipchat-go/hipchat"
)

type HipChatUserInformation struct {
	ID       string
	Name     string
	PhotoURL string
	Client   string
	Timezone string
	Title    string
}

type HipChatNetwork struct {
	client    *hipchat.Client
	rooms     map[string]string
	users     map[string]string
	apiClient *api.Client
	messages  chan chat.InMsg
	botName   string
}

func HipChatConnect(botID string, botPswd string, botName string, v2Token string) *HipChatNetwork {
retry:
	connected := make(chan *HipChatNetwork)
	go func() {
		fmt.Println("Attempting to connect as", botName)
		connected <- hipChatConnect(botID, botPswd, botName, v2Token)
	}()
	select {
	case conn := <-connected:
		if conn != nil {
			fmt.Println("Connected")
			return conn
		}
		fmt.Println("Retrying in 3 seconds")
		time.Sleep(3 * time.Second)
		goto retry
	case <-time.After(5 * time.Second):
		fmt.Println("Retrying in 10 seconds")
		time.Sleep(10 * time.Second)
		goto retry
	}
	return nil
}
func hipChatConnect(botID string, botPswd string, botName string, v2Token string) *HipChatNetwork {

	apiClient := api.NewClient(v2Token)

	client, err := hipchat.NewClient(botID, botPswd, "bot")
	if err != nil {
		log.Fatalln("Cannot create client:", err)
	}

	var botXMPJID, botMentionName *string
	for _, u := range client.Users() {
		if u.Name == botName {
			jid, name := u.Id, u.MentionName
			botXMPJID = &jid
			botMentionName = &name
		}
	}

	if botXMPJID == nil || botMentionName == nil {
		fmt.Println("Couldn't find bot jid or mention name")
		time.Sleep(10 * time.Second)
		return nil
	}
	go func() {
		for range client.OnConnect() {
			client.Status("chat")
		}
	}()

	go func() {
		client.KeepAlive()
	}()

	hipchatChatNetwork := &HipChatNetwork{client, make(map[string]string, 0), make(map[string]string, 0), apiClient, make(chan chat.InMsg, 0), botName}

	go func() {
		for x := range client.Messages() {
			func(m *hipchat.Message) {
				start := time.Now()
				// When happybot sends via the api messages come in the sender being the recipient but there is no /<name>. xmpp :(
				isPMFromHappyBot := !strings.Contains(m.From, "/")
				if len(m.From) == 0 || strings.HasPrefix(m.From, *botXMPJID) || strings.HasSuffix(m.From, botName) || isPMFromHappyBot {
					return
				}
				fromXMPJID := strings.Split(m.From, "/")[0]
				defer func() {
					if r := recover(); r != nil {
						hipchatChatNetwork.client.Say(fromXMPJID, botName, fmt.Sprintf("Unable to process '%v' due to error: %v", m.Body, r))
					}
				}()
				inMsg := chat.InMsg{}
				if m.Type == "groupchat" {
					// Because xmpp is super lame the id of the user who sent the message is not sent in the message
					// when in a room. So we go fetch it from the hipchat api.
					roomName, err := hipchatChatNetwork.roomNameFromXMPJID(fromXMPJID)
					panicErr(err)
					inMsg.RoomID = &roomName

					request, err := apiClient.NewRequest("GET", "room/"+roomName+"/history/"+m.ID+"?expand=message.from", nil)
					panicErr(err)

					var msgs map[string]api.Message
					_, err = apiClient.Do(request, &msgs)
					panicErr(err)

					if msg, ok := msgs["message"]; ok && msg.From != nil {
						userXMPJID := msg.From.(map[string]interface{})["xmpp_jid"].(string)
						mentionName, err := hipchatChatNetwork.mentionNameFromXMPJID(userXMPJID)
						panicErr(err)
						inMsg.From = mentionName
					} else {
						if !ok {
							panic(fmt.Sprintf("missing message response key(%v)", msgs))
						}
						return
					}
				} else if m.Type == "chat" {
					mentionName, err := hipchatChatNetwork.mentionNameFromXMPJID(fromXMPJID)
					panicErr(err)
					inMsg.From = mentionName
				}
				inMsg.ID = m.ID
				inMsg.Body = m.Body
				inMsg.ArrivedAt = start
				hipchatChatNetwork.messages <- inMsg
			}(x)
		}
	}()
	return hipchatChatNetwork
}

func (h *HipChatNetwork) NickName() string {
	return h.botName
}
func (h *HipChatNetwork) roomNameFromXMPJID(id string) (string, error) {
	name, ok := h.rooms[id]
	if ok {
		return name, nil
	}
	for _, r := range h.client.Rooms() {
		h.rooms[r.Id] = r.Name
	}
	name, ok = h.rooms[id]
	if ok {
		return name, nil
	}

	return "", fmt.Errorf("Unable to find room with xmpjid:", id)
}

func (h *HipChatNetwork) mentionNameFromXMPJID(id string) (string, error) {
	name, ok := h.users[id]
	if ok {
		return name, nil
	}
	for _, u := range h.client.Users() {
		h.users[u.Id] = u.MentionName
	}
	name, ok = h.users[id]
	if ok {
		return name, nil
	}

	return "", fmt.Errorf("Unable to find user with xmpjid:", id)
}

func (h *HipChatNetwork) SendPM(m chat.OutMsg) error {
	for id, name := range h.users {
		if name == m.To {
			h.client.Say(id, h.botName, m.Body)
			return nil
		}
	}
	panic(fmt.Sprintln("pm", m))
	return fmt.Errorf("Unable to find user id with name %v", m.To)
}

func (h *HipChatNetwork) Send(m chat.OutMsg) error {
	for id, name := range h.rooms {
		if name == m.To {
			h.client.Say(id, h.botName, m.Body)
			return nil
		}
	}
	panic(fmt.Sprintln("room", m))
	return fmt.Errorf("Unable to find room id with room name %v", m.To)
}

func (h *HipChatNetwork) JoinRoom(id string) error {
	h.client.Join(id+"@conf.hipchat.com", h.botName)
	return nil
}

func (h *HipChatNetwork) SetStatus(s string) error {
	return nil
}

func (h *HipChatNetwork) Messages() <-chan chat.InMsg {
	return h.messages
}
