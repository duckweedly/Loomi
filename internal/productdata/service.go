package productdata

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
)

type Service interface {
	CurrentIdentity(context.Context, identity.LocalIdentity) (User, error)
	CreateThread(context.Context, identity.LocalIdentity, CreateThreadInput) (Thread, error)
	ListThreads(context.Context, identity.LocalIdentity, bool) ([]Thread, error)
	GetThread(context.Context, identity.LocalIdentity, string) (Thread, error)
	UpdateThread(context.Context, identity.LocalIdentity, string, UpdateThreadInput) (Thread, error)
	ArchiveThread(context.Context, identity.LocalIdentity, string) (Thread, error)
	CreateMessage(context.Context, identity.LocalIdentity, string, CreateMessageInput) (Message, bool, error)
	ListMessages(context.Context, identity.LocalIdentity, string) ([]Message, error)
}

type SeedService interface {
	Service
	UpsertSeedThread(context.Context, identity.LocalIdentity, SeedThreadInput) (Thread, error)
	UpsertSeedMessage(context.Context, identity.LocalIdentity, SeedMessageInput) (Message, error)
}

type Repository interface {
	SeedService
}

type MemoryService struct {
	mu       sync.Mutex
	now      func() time.Time
	users    map[string]User
	threads  map[string]Thread
	messages map[string][]Message
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		now:      time.Now,
		users:    map[string]User{},
		threads:  map[string]Thread{},
		messages: map[string][]Message{},
	}
}

func (s *MemoryService) CurrentIdentity(_ context.Context, ident identity.LocalIdentity) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ensureUserLocked(ident), nil
}

func (s *MemoryService) CreateThread(_ context.Context, ident identity.LocalIdentity, input CreateThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	return s.upsertThreadLocked(NewThreadID(), user.ID, title, input.Mode), nil
}

func (s *MemoryService) UpsertSeedThread(_ context.Context, ident identity.LocalIdentity, input SeedThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	return s.upsertThreadLocked(input.ID, user.ID, title, input.Mode), nil
}

func (s *MemoryService) ListThreads(_ context.Context, ident identity.LocalIdentity, includeArchived bool) ([]Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	threads := make([]Thread, 0, len(s.threads))
	for _, thread := range s.threads {
		if thread.UserID != user.ID {
			continue
		}
		if !includeArchived && thread.LifecycleStatus == ThreadLifecycleArchived {
			continue
		}
		threads = append(threads, thread)
	}
	sort.SliceStable(threads, func(i, j int) bool { return threads[i].UpdatedAt.After(threads[j].UpdatedAt) })
	return threads, nil
}

func (s *MemoryService) GetThread(_ context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	return thread, nil
}

func (s *MemoryService) UpdateThread(_ context.Context, ident identity.LocalIdentity, threadID string, input UpdateThreadInput) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if input.Title != nil {
		title, err := NormalizeThreadTitle(*input.Title)
		if err != nil {
			return Thread{}, err
		}
		thread.Title = title
	}
	if input.Mode != nil {
		if err := ValidateThreadMode(*input.Mode); err != nil {
			return Thread{}, err
		}
		thread.Mode = *input.Mode
	}
	thread.UpdatedAt = s.now()
	s.threads[thread.ID] = thread
	return thread, nil
}

func (s *MemoryService) ArchiveThread(_ context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	now := s.now()
	thread.LifecycleStatus = ThreadLifecycleArchived
	thread.ArchivedAt = &now
	thread.UpdatedAt = now
	s.threads[thread.ID] = thread
	return thread, nil
}

func (s *MemoryService) CreateMessage(_ context.Context, ident identity.LocalIdentity, threadID string, input CreateMessageInput) (Message, bool, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, false, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	message, created, err := s.upsertMessageLocked(NewMessageID(), threadID, user.ID, content, clientMessageID)
	return message, created, err
}

func (s *MemoryService) UpsertSeedMessage(_ context.Context, ident identity.LocalIdentity, input SeedMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	message, _, err := s.upsertMessageLocked(input.ID, input.ThreadID, user.ID, content, clientMessageID)
	return message, err
}

func (s *MemoryService) ListMessages(_ context.Context, ident identity.LocalIdentity, threadID string) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return nil, NewError(CodeThreadNotFound, "Thread not found.")
	}
	messages := append([]Message(nil), s.messages[threadID]...)
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].CreatedAt.Before(messages[j].CreatedAt) })
	return messages, nil
}

func (s *MemoryService) ensureUserLocked(ident identity.LocalIdentity) User {
	if user, ok := s.users[ident.UserID]; ok {
		return user
	}
	now := s.now()
	user := User{ID: ident.UserID, DisplayName: ident.DisplayName, CreatedAt: now, UpdatedAt: now}
	s.users[user.ID] = user
	return user
}

func (s *MemoryService) upsertThreadLocked(id string, userID string, title string, mode ThreadMode) Thread {
	now := s.now()
	thread, ok := s.threads[id]
	if !ok {
		thread = Thread{ID: id, UserID: userID, CreatedAt: now}
	}
	thread.Title = title
	thread.Mode = mode
	thread.LifecycleStatus = ThreadLifecycleActive
	thread.ArchivedAt = nil
	thread.UpdatedAt = now
	s.threads[thread.ID] = thread
	return thread
}

func (s *MemoryService) upsertMessageLocked(id string, threadID string, userID string, content string, clientMessageID *string) (Message, bool, error) {
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != userID {
		return Message{}, false, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if clientMessageID != nil {
		for _, message := range s.messages[threadID] {
			if message.UserID == userID && message.ClientMessageID != nil && *message.ClientMessageID == *clientMessageID {
				return message, false, nil
			}
		}
	}
	now := s.now()
	message := Message{ID: id, ThreadID: threadID, UserID: userID, Role: MessageRoleUser, Content: content, Metadata: map[string]any{}, ClientMessageID: clientMessageID, CreatedAt: now}
	s.messages[threadID] = append(s.messages[threadID], message)
	thread.UpdatedAt = now
	s.threads[threadID] = thread
	return message, true, nil
}
