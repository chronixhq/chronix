package cxrestapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	eventspkg "chronix/internal/events"
	"chronix/internal/secret"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

var (
	authKeys = map[string]struct{}{}
)

// Auth Keys
func RevokeAuthKey(key string) {
	if len(key) > 0 {
		delete(authKeys, key)
		eventspkg.ShutdownAuthkeySession(key)
	}
}

func SyncAuthKeys() error {
	users, err := db.CxUser.Where(db.CxUser.Enabled.Is(true), db.CxUser.Suspended.Is(false)).Select(db.CxUser.ID, db.CxUser.Sv).Find()
	if err != nil {
		return err
	}
	activeUsers := make(map[int64]int32)
	for _, u := range users {
		activeUsers[u.ID] = u.Sv
	}

	keysToDelete := []string{}
	for k := range authKeys {
		parts := strings.Split(k, ":")
		if len(parts) < 2 {
			continue
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		sv, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			continue
		}

		if activeSv, ok := activeUsers[id]; !ok || activeSv != int32(sv) {
			keysToDelete = append(keysToDelete, k)
		}
	}

	if len(keysToDelete) > 0 {
		for _, k := range keysToDelete {
			delete(authKeys, k)
			eventspkg.ShutdownAuthkeySession(k)
		}
	}
	return SaveAuthKeys()
}

func RevokeUserAuthKeys(userID int64) {
	for k := range authKeys {
		parts := strings.Split(k, ":")
		if len(parts) < 1 {
			continue
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		if id == userID {
			delete(authKeys, k)
		}
	}
	eventspkg.ShutdownUserSSESessions(userID)
}

func SetupAuthKeys() {
	err := LoadAuthKeys()
	if err != nil {
		slog.Error("Error loading auth keys", "error", err)
	}
	go func() {
		for {
			time.Sleep(time.Minute * 1)
			err := SyncAuthKeys()
			if err != nil {
				slog.Error("Error syncing auth keys", "error", err)
			}
		}
	}()
}

func LoadAuthKeys() error {
	keys, err := db.AuthKey.Find()
	if err != nil {
		return err
	}
	authKeys = make(map[string]struct{})
	for _, k := range keys {
		if k.AuthKey != nil {
			dec, _ := secret.Decrypt(*k.AuthKey)
			authKeys[dec] = struct{}{}
		}
	}
	return nil
}

func SaveAuthKeys() error {
	keys := []*models.AuthKey{}
	for k := range authKeys {
		enc, _ := secret.Encrypt(k)
		keys = append(keys, &models.AuthKey{AuthKey: &enc})
	}
	_, err := db.AuthKey.Where(db.AuthKey.AuthKey.IsNotNull()).Delete()
	if err != nil {
		return err
	}
	return db.AuthKey.Save(keys...)
}
