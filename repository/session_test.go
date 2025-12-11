package repository

import (
	"sync"
	"testing"
)

func TestSessionRepository_SetAndGet(t *testing.T) {
	repo := NewSessionRepository()

	sessionID := "test-session-id"
	username := "testuser"

	repo.Set(sessionID, username)

	retrievedUsername, ok := repo.Get(sessionID)
	if !ok {
		t.Errorf("Get() returned ok = false, want true")
	}
	if retrievedUsername != username {
		t.Errorf("Get() username = %v, want %v", retrievedUsername, username)
	}

	_, ok = repo.Get("nonexistent-session")
	if ok {
		t.Errorf("Get() for nonexistent session returned ok = true, want false")
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	repo := NewSessionRepository()

	sessionID := "test-session-id"
	username := "testuser"

	repo.Set(sessionID, username)

	_, ok := repo.Get(sessionID)
	if !ok {
		t.Fatalf("Session should exist before delete")
	}

	repo.Delete(sessionID)

	_, ok = repo.Get(sessionID)
	if ok {
		t.Errorf("Session should not exist after delete")
	}
}

func TestSessionRepository_ConcurrentAccess(t *testing.T) {
	repo := NewSessionRepository()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := range numGoroutines {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			sessionID := "session-" + string(rune(id))
			username := "user-" + string(rune(id))
			repo.Set(sessionID, username)
		}(i)

		go func(id int) {
			defer wg.Done()
			sessionID := "session-" + string(rune(id))
			repo.Get(sessionID)
		}(i)

		go func(id int) {
			defer wg.Done()
			sessionID := "session-" + string(rune(id))
			repo.Delete(sessionID)
		}(i)
	}

	wg.Wait()
}

func TestSessionRepository_UpdateSession(t *testing.T) {
	repo := NewSessionRepository()

	sessionID := "test-session-id"
	username1 := "user1"
	username2 := "user2"

	repo.Set(sessionID, username1)

	retrievedUsername, ok := repo.Get(sessionID)
	if !ok {
		t.Fatalf("Session should exist")
	}
	if retrievedUsername != username1 {
		t.Errorf("Get() username = %v, want %v", retrievedUsername, username1)
	}

	repo.Set(sessionID, username2)

	retrievedUsername, ok = repo.Get(sessionID)
	if !ok {
		t.Fatalf("Session should still exist")
	}
	if retrievedUsername != username2 {
		t.Errorf("Get() username = %v, want %v", retrievedUsername, username2)
	}
}

func TestSessionRepository_MultipleSessionsPerUser(t *testing.T) {
	repo := NewSessionRepository()

	username := "testuser"
	sessionID1 := "session-1"
	sessionID2 := "session-2"

	repo.Set(sessionID1, username)
	repo.Set(sessionID2, username)

	user1, ok1 := repo.Get(sessionID1)
	user2, ok2 := repo.Get(sessionID2)

	if !ok1 || !ok2 {
		t.Errorf("Both sessions should exist")
	}
	if user1 != username || user2 != username {
		t.Errorf("Both sessions should return the same username")
	}

	repo.Delete(sessionID1)

	_, ok1 = repo.Get(sessionID1)
	user2, ok2 = repo.Get(sessionID2)

	if ok1 {
		t.Errorf("Session 1 should not exist after delete")
	}
	if !ok2 {
		t.Errorf("Session 2 should still exist")
	}
	if user2 != username {
		t.Errorf("Session 2 should still return the correct username")
	}
}
