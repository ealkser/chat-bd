package services

import (
	"chat-bd/repositories"
)

type ChatService struct {
	repo *repositories.ChatRepository
}

func NewChatService(repo *repositories.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) GetChatsForUser(userID int64) ([]repositories.ChatResponse, error) {
	chats, err := s.repo.GetChatsForUser(userID)
	if err != nil {
		return nil, err
	}

	var responses []repositories.ChatResponse
	for _, c := range chats {
		responses = append(responses, repositories.ChatResponse{
			ID:            c.ID,
			WithUserName:  c.WithUserName,
			LastMessage:   c.LastMessage,
			LastMessageAt: c.LastMessageAt,
		})
	}

	return responses, nil
}

func (s *ChatService) GetMessages(chatID int64) ([]repositories.MessageResponse, error) {
	messages, err := s.repo.GetMessages(chatID)
	if err != nil {
		return nil, err
	}

	var responses []repositories.MessageResponse
	for _, m := range messages {
		responses = append(responses, repositories.MessageResponse{
			ID:        m.ID,
			SenderID:  m.SenderID,
			Content:   m.Content,
			Read:      m.Read,
			CreatedAt: m.CreatedAt,
		})
	}

	return responses, nil
}

func (s *ChatService) SendMessage(chatID, senderID int64, content string) (int64, error) {
	return s.repo.CreateMessage(chatID, senderID, content)
}

func (s *ChatService) StartChat(currentUser, targetUser int64) (int64, error) {
	return s.repo.GetOrCreateChat(currentUser, targetUser)
}

func (s *ChatService) UserExists(email string) (int64, bool, error) {
	return s.repo.UserExists(email)
}

func (s *ChatService) IsUserInChat(chatID, userID int64) (bool, error) {
	return s.repo.IsUserInChat(chatID, userID)
}

func (s *ChatService) GetChatIDsForUser(userID int64) ([]int64, error) {
	return s.repo.GetChatIDsForUser(userID)
}
