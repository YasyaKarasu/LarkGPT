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

func p2pChat(messageevent *MessageEvent) {
	chat := chatgpt.New(config.C.ChatGPT.OpenaiKey, "", 60*time.Second)
	defer chat.Close()

	openid := messageevent.Sender.Sender_id.Open_id
	if session.GetContextSession(openid) != "" {
		err := chat.ChatContext.LoadConversation(openid)
		if err != nil {
			logrus.WithField("openid", openid).Error(err)
		}
	}

	answer, err := chat.ChatWithContext(messageevent.Message.Content)
	if err != nil {
		logrus.Error(err)
		if strings.Contains(fmt.Sprintf("%v", err), "maximum text length exceeded") {
			session.ClearContextSession(openid)
			global.Cli.MessageSend(
				feishuapi.UserOpenId,
				openid,
				feishuapi.Text,
				fmt.Sprintf("请求openai失败了，错误信息：%v，看起来是超过最大对话限制了，已自动重置您的对话", err.Error()),
			)
			p2pChat(messageevent)
		} else {
			global.Cli.MessageSend(
				feishuapi.UserOpenId,
				openid,
				feishuapi.Text,
				fmt.Sprintf("请求openai失败了，错误信息：%v", err.Error()),
			)
		}
		return
	}

	answer = strings.TrimSpace(answer)
	answer = strings.Trim(answer, "\n")
	global.Cli.MessageSend(feishuapi.UserOpenId, openid, feishuapi.Text, answer)

	chat.ChatContext.SaveConversation(openid)
}
