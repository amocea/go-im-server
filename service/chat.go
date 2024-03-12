package service

import "github.com/amocea/go-im-chat/defs"

type ChatService struct {
	Service
}

// CheckToken 验证 token 是否合法
func (s ChatService) CheckToken(id int64, token string) bool {
	var cond = defs.User{
		Id:    id,
		Token: token,
	}
	exist, err := s.DB.Exist(&cond)
	if err != nil {
		return false
	}
	return exist
}
