package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HistoryManager struct {
	conversations map[string]*Conversation // Key: ConversationID
	filePath      string
	mu            sync.RWMutex
}

func NewHistoryManager(filePath string) (*HistoryManager, error) {
	hm := &HistoryManager{
		conversations: make(map[string]*Conversation),
		filePath:      filePath,
	}
	if err := hm.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return hm, nil
}

func (hm *HistoryManager) load() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	file, err := os.ReadFile(hm.filePath)
	if err != nil {
		return err
	}

	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, &hm.conversations)
}

func (hm *HistoryManager) save() error {
	data, err := json.MarshalIndent(hm.conversations, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(hm.filePath, data, 0644)
}

func (hm *HistoryManager) CreateConversation(userID, title string) (*Conversation, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	id := generateID()
	if title == "" {
		title = "New Chat"
	}

	conv := &Conversation{
		ID:        id,
		UserID:    userID,
		Title:     title,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	hm.conversations[id] = conv
	if err := hm.save(); err != nil {
		return nil, err
	}
	return conv, nil
}

func (hm *HistoryManager) AddMessage(conversationID, role, content string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	conv, exists := hm.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	msg := Message{
		ID:        generateID(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()

	return hm.save()
}

// GetConversations returns a list of conversations for a user (without messages to save bandwidth)
func (hm *HistoryManager) GetConversations(userID string) ([]Conversation, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var result []Conversation
	for _, conv := range hm.conversations {
		if conv.UserID == userID {
			// Create a copy without messages for the list view
			c := *conv
			c.Messages = nil
			result = append(result, c)
		}
	}

	// Sort by UpdatedAt desc
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	return result, nil
}

func (hm *HistoryManager) GetConversation(conversationID string) (*Conversation, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	conv, exists := hm.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	// Return copy
	c := *conv
	messages := make([]Message, len(conv.Messages))
	copy(messages, conv.Messages)
	c.Messages = messages
	return &c, nil
}

func (hm *HistoryManager) DeleteConversation(conversationID string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.conversations[conversationID]; !exists {
		return fmt.Errorf("conversation not found")
	}

	delete(hm.conversations, conversationID)
	return hm.save()
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
