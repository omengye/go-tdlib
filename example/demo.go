package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/omengye/go-tdlib/client"
)

var (
	ChatsInfo = make(map[int64]string)
)

func main() {
	var (
		apiIdRaw = os.Getenv("API_ID")
		apiHash  = os.Getenv("API_HASH")
	)

	apiId64, err := strconv.ParseInt(apiIdRaw, 10, 32)
	if err != nil {
		log.Fatalf("strconv.Atoi error: %s", err)
	}

	apiId := int32(apiId64)

	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               apiId,
		ApiHash:             apiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}
	authorizer := client.ClientAuthorizer(tdlibParameters)
	go client.CliInteractor(authorizer)

	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 0,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %s", err)
	}

	proxy := client.WithProxy(&client.AddProxyRequest{
		Server: "127.0.0.1",
		Port:   11034,
		Enable: true,
		Type:   &client.ProxyTypeSocks5{},
	})

	// ----------------------- handle chat message -----------------------
	resHandCallback := func(result client.Type) {
		switch result.GetConstructor() {
		case client.ConstructorUpdateChatLastMessage:
			res := result.(*client.UpdateChatLastMessage)
			if res.LastMessage == nil || ChatsInfo[res.ChatId] == "" {
				return
			}
			chatName := ChatsInfo[res.ChatId]
			switch res.LastMessage.Content.MessageContentConstructor() {
			case client.ConstructorMessageText:
				content := res.LastMessage.Content.(*client.MessageText)
				log.Printf("Channel: %s, Content: %s", chatName, content.Text.Text)
			}
		}
	}

	tdlibClient, err := client.NewClient(authorizer, client.WithResultHandler(client.NewCallbackResultHandler(resHandCallback)), proxy)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s (commit: %s)", versionOption.(*client.OptionValueString).Value, commitOption.(*client.OptionValueString).Value)

	if commitOption.(*client.OptionValueString).Value != client.TDLIB_VERSION {
		log.Printf("TDLib version supported by the library (%s) is not the same as TDLib version (%s)", client.TDLIB_VERSION, commitOption.(*client.OptionValueString).Value)
	}

	// ----------------------- login user info -----------------------

	me, err := tdlibClient.GetMe(context.Background())
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("Me: %s %s, Id: %d", me.FirstName, me.LastName, me.Id)

	userInfo, err := tdlibClient.GetUserFullInfo(context.Background(), &client.GetUserFullInfoRequest{
		UserId: me.Id,
	})
	if err != nil {
		log.Fatalf("GetUserFullInfo error: %s", err)
	}

	// ----------------------- all chats -----------------------
	chats, err := tdlibClient.GetChats(context.Background(), &client.GetChatsRequest{
		Limit: 100,
	})
	if err != nil {
		log.Fatalf("GetChats error: %s", err)
	}
	for _, id := range chats.ChatIds {
		if c, err := tdlibClient.GetChat(context.Background(), &client.GetChatRequest{
			ChatId: id,
		}); err == nil {
			log.Printf("chat title: %s, id: %d", c.Title, c.Id)

			// find all channels
			switch c.Type.ChatTypeConstructor() {
			case client.ConstructorChatTypeSupergroup:
				ct := c.Type.(*client.ChatTypeSupergroup)
				if ct.IsChannel {
					ChatsInfo[c.Id] = c.Title
				}
			}
		}
	}

	// ----------------------- personal channel -----------------------

	chatId := userInfo.PersonalChatId
	chat, err := tdlibClient.GetChat(context.Background(), &client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		log.Fatalf("PersonalChat error: %s", err)
	}
	log.Printf("chat title: %s", chat.Title)

	// ----------------------- bots -----------------------

	bots, err := tdlibClient.GetOwnedBots(context.Background())
	if err != nil {
		log.Fatalf("GetOwnedBots error: %s", err)
	}
	for _, botId := range bots.UserIds {
		bot, err := tdlibClient.GetUser(context.Background(), &client.GetUserRequest{
			UserId: botId,
		})
		if err != nil {
			log.Fatalf("GetUser Bot error: %s", err)
		}
		log.Printf("bot username: %s, id: %d", bot.Usernames.EditableUsername, bot.Id)
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	tdlibClient.Close(context.Background())
	os.Exit(1)
}
