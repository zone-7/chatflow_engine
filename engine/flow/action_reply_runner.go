package flow

import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

func init() {
	andflow.RegistActionRunner("reply", &ReplyRunner{})
}

type ReplyRunner struct {
	BaseRunner
}

func (r *ReplyRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.getChatSession(s)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content := prop["reply_template"]

	if len(content) > 0 {

		uid, _ := uuid.NewV4()
		mid := strings.ReplaceAll(uid.String(), "-", "")
		chatSession.Response(meta.ChatFlowMessage{MessageId: mid, Content: content, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)

	}
	return andflow.RESULT_SUCCESS, err
}
