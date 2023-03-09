package receiveMessage

import (
	"fmt"
	"strings"
	"time"
	"xlab-feishu-robot/config"
	"xlab-feishu-robot/pkg/chatgpt"
	"xlab-feishu-robot/pkg/global"
	"xlab-feishu-robot/pkg/session"

	"github.com/YasyaKarasu/feishuapi"
	"github.com/sirupsen/logrus"
)

func groupChat(messageevent *MessageEvent) {
	chat := chatgpt.New(config.C.ChatGPT.OpenaiKey, "", 60*time.Second)
	defer chat.Close()

	chatid := messageevent.Message.Chat_id
	if session.GetContextSession(chatid) != "" {
		err := chat.ChatContext.LoadConversation(chatid)
		if err != nil {
			logrus.WithField("openid", chatid).Error(err)
		}
	}

	answer, err := chat.ChatWithContext(messageevent.Message.Content)
	if err != nil {
		logrus.Error(err)
		if strings.Contains(fmt.Sprintf("%v", err), "maximum text length exceeded") {
			session.ClearContextSession(chatid)
			global.Cli.MessageSend(
				feishuapi.GroupChatId,
				chatid,
				feishuapi.Text,
				fmt.Sprintf("请求openai失败了，错误信息：%v，看起来是超过最大对话长度限制了，已自动重置您的对话", err.Error()),
			)
			groupChat(messageevent)
		} else {
			global.Cli.MessageSend(
				feishuapi.GroupChatId,
				chatid,
				feishuapi.Text,
				fmt.Sprintf("请求openai失败了，错误信息：%v", err.Error()),
			)
		}
		return
	}

	answer = strings.TrimSpace(answer)
	answer = strings.Trim(answer, "\n")
	global.Cli.MessageSend(feishuapi.GroupChatId, messageevent.Message.Chat_id, feishuapi.Text, answer)

	chat.ChatContext.SaveConversation(chatid)
}
