package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/orca-ng/orca/internal/crypto"
	"github.com/orca-ng/orca/internal/cyberark"
	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/pipeline"
	"github.com/orca-ng/orca/internal/services"
)

// UserSyncHandler handles user synchronization operations
type UserSyncHandler struct {
	db          *database.GormDB
	logger      *logrus.Logger
	certManager *services.CertificateManager
	encryptor   *crypto.Encryptor
}

// NewUserSyncHandler creates a new user sync handler
func NewUserSyncHandler(db *database.GormDB, logger *logrus.Logger, certManager *services.CertificateManager, encryptor *crypto.Encryptor) *UserSyncHandler {
	return &UserSyncHandler{
		db:          db,
		logger:      logger,
		certManager: certManager,
		encryptor:   encryptor,
	}
}

// UserSyncPayload represents the payload for user sync operations
type UserSyncPayload struct {
	InstanceID string `json:"instance_id"`
	SyncMode   string `json:"sync_mode"` // "manual" or "scheduled"
	PageSize   *int   `json:"page_size,omitempty"` // override instance default
}

// UserSyncResult represents the result of a user sync operation
type UserSyncResult struct {
	TotalUsers     int       `json:"total_users"`
	ProcessedUsers int       `json:"processed_users"`
	NewUsers       int       `json:"new_users"`
	UpdatedUsers   int       `json:"updated_users"`
	DeletedUsers   int       `json:"deleted_users"`
	Errors         []string  `json:"errors,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
}

// Handle processes the user sync operation
func (h *UserSyncHandler) Handle(ctx context.Context, op *pipeline.Operation) error {
	startTime := time.Now()
	h.logger.WithField("operation_id", op.ID).Info("Starting user sync operation")

	// Parse payload
	var payload UserSyncPayload
	if err := json.Unmarshal(op.Payload, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	// Load CyberArk instance
	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ?", payload.InstanceID).Error; err != nil {
		return fmt.Errorf("load instance: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"instance_name": instance.Name,
	}).Debug("Loaded CyberArk instance for sync")

	// Update sync status to running
	h.updateInstanceSyncStatus(&instance, "running", nil)

	// Get CyberArk client from context or create new one
	var client *cyberark.Client
	if ctxClient := ctx.Value("cyberark_client"); ctxClient != nil {
		client = ctxClient.(*cyberark.Client)
		h.logger.Debug("Using CyberArk client from context")
	} else {
		// Create new client
		h.logger.WithFields(logrus.Fields{
			"instance_id": instance.ID,
			"encrypted_password_length": len(instance.PasswordEncrypted),
			"encrypted_password_sample": instance.PasswordEncrypted[:10] + "...",
		}).Debug("Attempting to decrypt password")
		
		password, err := h.encryptor.Decrypt(instance.PasswordEncrypted)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"encrypted_length": len(instance.PasswordEncrypted),
			}).Error("Failed to decrypt password")
			h.updateInstanceSyncStatus(&instance, "failed", err)
			return fmt.Errorf("decrypt password: %w", err)
		}

		client, err = cyberark.NewClient(cyberark.Config{
			BaseURL:        instance.BaseURL,
			Username:       instance.Username,
			Password:       password,
			SkipTLSVerify:  instance.SkipTLSVerify,
			RequestTimeout: 30 * time.Second,
			CertManager:    h.certManager,
		})
		if err != nil {
			h.updateInstanceSyncStatus(&instance, "failed", err)
			return fmt.Errorf("create client: %w", err)
		}

		// Authenticate if not already authenticated
		if !client.IsAuthenticated() {
			if _, err := client.Authenticate(); err != nil {
				h.updateInstanceSyncStatus(&instance, "failed", err)
				return fmt.Errorf("authenticate: %w", err)
			}
		}
	}

	// Determine page size
	// Get page size from sync config or use default
	pageSize := 100
	if payload.PageSize != nil && *payload.PageSize > 0 {
		pageSize = *payload.PageSize
	}

	// Perform the sync
	result, err := h.syncUsers(ctx, client, &instance, pageSize)
	if err != nil {
		h.updateInstanceSyncStatus(&instance, "failed", err)
		return fmt.Errorf("sync users: %w", err)
	}

	// Update operation result
	resultBytes, _ := json.Marshal(result)
	resultRaw := json.RawMessage(resultBytes)
	op.Result = &resultRaw

	// Update instance sync status
	h.updateInstanceSyncStatus(&instance, "success", nil)

	h.logger.WithFields(map[string]interface{}{
		"operation_id":    op.ID,
		"instance_id":     instance.ID,
		"total_users":     result.TotalUsers,
		"processed_users": result.ProcessedUsers,
		"new_users":       result.NewUsers,
		"updated_users":   result.UpdatedUsers,
		"deleted_users":   result.DeletedUsers,
		"duration":        time.Since(startTime).Seconds(),
	}).Info("User sync operation completed")

	return nil
}

// syncUsers performs the actual synchronization
func (h *UserSyncHandler) syncUsers(ctx context.Context, client *cyberark.Client, instance *gormmodels.CyberArkInstance, pageSize int) (*UserSyncResult, error) {
	result := &UserSyncResult{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	// Track users and memberships seen in this sync
	seenUserIDs := make(map[string]bool)
	seenMembershipKeys := make(map[string]bool) // key format: "userID:groupID"
	pageOffset := 1 // CyberArk pagination starts at 1
	maxRetries := 3
	
	for {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Fetch users page with retry logic
		var listResp *cyberark.UserListResponse
		var err error
		
		for attempt := 0; attempt <= maxRetries; attempt++ {
			h.logger.WithFields(map[string]interface{}{
				"page_offset": pageOffset,
				"page_size":   pageSize,
				"attempt":     attempt,
			}).Debug("Fetching users page")

			listResp, err = client.ListUsers(ctx, cyberark.ListUsersOptions{
				PageOffset:      pageOffset,
				PageSize:        pageSize,
				ExtendedDetails: true,
			})
			
			if err == nil {
				break // Success
			}
			
			// Check if error is retryable
			if !h.isRetryableError(err) || attempt == maxRetries {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch users at page %d: %v", pageOffset, err))
				return result, err
			}
			
			// Handle token expiration mid-sync
			if h.isAuthError(err) {
				h.logger.Info("Token expired during sync, attempting to re-authenticate")
				if _, authErr := client.AuthenticateWithContext(ctx); authErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to re-authenticate: %v", authErr))
					return result, authErr
				}
			}
			
			// Wait before retry
			waitTime := time.Duration(attempt+1) * time.Second
			h.logger.WithField("wait_seconds", waitTime.Seconds()).Debug("Waiting before retry")
			time.Sleep(waitTime)
		}

		// Check if page is empty (indicates end of results)
		if len(listResp.Users) == 0 {
			h.logger.Debug("Received empty user page, ending pagination")
			break
		}

		// Process users in this page
		for _, caUser := range listResp.Users {
			if err := h.processUser(instance, &caUser, seenUserIDs, seenMembershipKeys, result); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to process user %s: %v", caUser.Username, err))
				h.logger.WithError(err).WithField("username", caUser.Username).Error("Failed to process user")
			} else {
				result.ProcessedUsers++
			}
		}
		
		// Note: listResp.Total is the count in current page, not total users
		result.TotalUsers += len(listResp.Users)

		// Move to next page
		pageOffset++
	}

	// Mark users not seen in this sync as deleted
	if err := h.markDeletedUsers(instance.ID, seenUserIDs, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to mark deleted users: %v", err))
	}
	
	// Mark group memberships not seen in this sync as deleted
	if err := h.markDeletedMemberships(instance.ID, seenMembershipKeys, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to mark deleted memberships: %v", err))
	}

	result.CompletedAt = time.Now()
	return result, nil
}

// processUser processes a single user and their group memberships
func (h *UserSyncHandler) processUser(instance *gormmodels.CyberArkInstance, caUser *cyberark.User, seenUserIDs map[string]bool, seenMembershipKeys map[string]bool, result *UserSyncResult) error {
	userID := strconv.Itoa(caUser.ID)
	seenUserIDs[userID] = true
	
	h.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"user_id": userID,
		"username": caUser.Username,
	}).Debug("Processing user")

	// Check if user exists
	var existingUser gormmodels.CyberArkUser
	err := h.db.Where("cyberark_instance_id = ? AND user_id = ?", instance.ID, userID).First(&existingUser).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new user
		newUser := h.buildCyberArkUser(instance.ID, caUser)
		h.logger.WithFields(logrus.Fields{
			"cyberark_instance_id": newUser.CyberArkInstanceID,
			"username": newUser.Username,
			"user_id": newUser.UserID,
			"user_struct": fmt.Sprintf("%+v", newUser),
		}).Debug("Creating new user")
		
		// Enable debug mode temporarily
		tx := h.db.Debug().Create(newUser)
		if tx.Error != nil {
			return fmt.Errorf("create user: %w", err)
		}
		result.NewUsers++
	} else if err != nil {
		return fmt.Errorf("query user: %w", err)
	} else {
		// Update existing user
		updates := h.buildUserUpdates(caUser)
		updates["last_synced_at"] = time.Now()
		updates["is_deleted"] = false
		updates["deleted_at"] = nil
		
		if err := h.db.Model(&existingUser).Updates(updates).Error; err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		result.UpdatedUsers++
	}
	
	// Process group memberships
	if err := h.processUserGroupMemberships(instance, caUser, seenMembershipKeys); err != nil {
		return fmt.Errorf("process group memberships: %w", err)
	}
	
	// Process vault authorizations
	if err := h.processUserVaultAuthorizations(instance, caUser); err != nil {
		return fmt.Errorf("process vault authorizations: %w", err)
	}

	return nil
}

// buildCyberArkUser builds a CyberArkUser model from the API response
func (h *UserSyncHandler) buildCyberArkUser(instanceID string, caUser *cyberark.User) *gormmodels.CyberArkUser {
	user := &gormmodels.CyberArkUser{
		CyberArkInstanceID:    instanceID,
		Username:              caUser.Username,
		UserID:                strconv.Itoa(caUser.ID),
		UserType:              caUser.UserType,
		ComponentUser:         caUser.ComponentUser,
		Suspended:             caUser.Suspended,
		EnableUser:            caUser.EnableUser,
		LastSyncedAt:          time.Now(),
	}

	// Set location if not empty
	if caUser.Location != "" {
		user.Location = &caUser.Location
	}

	// Set personal details
	if caUser.PersonalDetails != nil {
		if caUser.PersonalDetails.FirstName != "" {
			user.FirstName = &caUser.PersonalDetails.FirstName
		}
		if caUser.PersonalDetails.LastName != "" {
			user.LastName = &caUser.PersonalDetails.LastName
		}
		user.ChangePassOnNextLogon = caUser.PersonalDetails.ChangePassOnNextLogon
		
		if caUser.PersonalDetails.ExpiryDate != nil {
			expiryTime := cyberark.TimestampToTime(caUser.PersonalDetails.ExpiryDate)
			user.ExpiryDate = expiryTime
		}
	}

	// Set email if available
	if caUser.Internet != nil && caUser.Internet.Email != "" {
		user.Email = &caUser.Internet.Email
	}

	// Set last login time
	if caUser.LastSuccessfulLoginAt != nil {
		loginTime := cyberark.TimestampToTime(caUser.LastSuccessfulLoginAt)
		user.LastSuccessfulLoginAt = loginTime
	}

	return user
}

// buildUserUpdates builds the update map for an existing user
func (h *UserSyncHandler) buildUserUpdates(caUser *cyberark.User) map[string]interface{} {
	updates := map[string]interface{}{
		"username":    caUser.Username,
		"user_type":   caUser.UserType,
		"suspended":   caUser.Suspended,
		"enable_user": caUser.EnableUser,
	}

	// Update location
	if caUser.Location != "" {
		updates["location"] = caUser.Location
	} else {
		updates["location"] = nil
	}

	// Update personal details
	if caUser.PersonalDetails != nil {
		if caUser.PersonalDetails.FirstName != "" {
			updates["first_name"] = caUser.PersonalDetails.FirstName
		} else {
			updates["first_name"] = nil
		}
		
		if caUser.PersonalDetails.LastName != "" {
			updates["last_name"] = caUser.PersonalDetails.LastName
		} else {
			updates["last_name"] = nil
		}
		
		updates["change_pass_on_next_logon"] = caUser.PersonalDetails.ChangePassOnNextLogon
		
		if caUser.PersonalDetails.ExpiryDate != nil {
			updates["expiry_date"] = cyberark.TimestampToTime(caUser.PersonalDetails.ExpiryDate)
		} else {
			updates["expiry_date"] = nil
		}
	}

	// Update email
	if caUser.Internet != nil && caUser.Internet.Email != "" {
		updates["email"] = caUser.Internet.Email
	} else {
		updates["email"] = nil
	}

	// Update last login time
	if caUser.LastSuccessfulLoginAt != nil {
		updates["last_successful_login_at"] = cyberark.TimestampToTime(caUser.LastSuccessfulLoginAt)
	} else {
		updates["last_successful_login_at"] = nil
	}

	return updates
}

// markDeletedUsers marks users not seen in the sync as deleted
func (h *UserSyncHandler) markDeletedUsers(instanceID string, seenUserIDs map[string]bool, result *UserSyncResult) error {
	// Build list of seen user IDs
	userIDs := make([]string, 0, len(seenUserIDs))
	for id := range seenUserIDs {
		userIDs = append(userIDs, id)
	}

	// Mark users not in the list as deleted
	now := time.Now()
	res := h.db.Model(&gormmodels.CyberArkUser{}).
		Where("cyberark_instance_id = ? AND is_deleted = false", instanceID).
		Where("user_id NOT IN ?", userIDs).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		})

	if res.Error != nil {
		return res.Error
	}

	result.DeletedUsers = int(res.RowsAffected)
	return nil
}

// updateInstanceSyncStatus updates the sync status on the instance
func (h *UserSyncHandler) updateInstanceSyncStatus(instance *gormmodels.CyberArkInstance, status string, err error) {
	now := time.Now()
	updates := map[string]interface{}{
		"last_user_sync_at":     now,
		"last_user_sync_status": status,
	}

	if err != nil {
		errMsg := err.Error()
		updates["last_user_sync_error"] = errMsg
	} else {
		updates["last_user_sync_error"] = nil
	}

	if dbErr := h.db.Model(instance).Updates(updates).Error; dbErr != nil {
		h.logger.WithError(dbErr).Error("Failed to update instance sync status")
	}
}

// CanRetry determines if an error is retryable
func (h *UserSyncHandler) CanRetry(err error) bool {
	// Retry on network errors or temporary failures
	// Don't retry on authentication errors or permission issues
	errMsg := err.Error()
	
	// Don't retry authentication errors
	if containsSubstring(errMsg, "authentication", "unauthorized", "forbidden", "permission") {
		return false
	}
	
	// Retry network and temporary errors
	if containsSubstring(errMsg, "timeout", "connection", "network", "temporary") {
		return true
	}
	
	return false
}

// processUserGroupMemberships processes the group memberships for a user
func (h *UserSyncHandler) processUserGroupMemberships(instance *gormmodels.CyberArkInstance, caUser *cyberark.User, seenMembershipKeys map[string]bool) error {
	userID := strconv.Itoa(caUser.ID)
	
	for _, groupMembership := range caUser.GroupsMembership {
		membershipKey := fmt.Sprintf("%s:%d", userID, groupMembership.GroupID)
		seenMembershipKeys[membershipKey] = true
		
		// Check if membership exists
		var existingMembership gormmodels.CyberArkGroupMembership
		err := h.db.Where("cyberark_instance_id = ? AND user_id = ? AND group_id = ?", 
			instance.ID, userID, groupMembership.GroupID).First(&existingMembership).Error
		
		if err == gorm.ErrRecordNotFound {
			// Create new membership
			newMembership := &gormmodels.CyberArkGroupMembership{
				CyberArkInstanceID: instance.ID,
				UserID:             userID,
				Username:           caUser.Username,
				GroupID:            groupMembership.GroupID,
				GroupName:          groupMembership.GroupName,
				GroupType:          groupMembership.GroupType,
				LastSyncedAt:       time.Now(),
			}
			
			if err := h.db.Create(newMembership).Error; err != nil {
				return fmt.Errorf("create membership: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("query membership: %w", err)
		} else {
			// Update existing membership
			updates := map[string]interface{}{
				"username":       caUser.Username,
				"group_name":     groupMembership.GroupName,
				"group_type":     groupMembership.GroupType,
				"last_synced_at": time.Now(),
				"is_deleted":     false,
				"deleted_at":     nil,
			}
			
			if err := h.db.Model(&existingMembership).Updates(updates).Error; err != nil {
				return fmt.Errorf("update membership: %w", err)
			}
		}
	}
	
	return nil
}

// markDeletedMemberships marks group memberships not seen in the sync as deleted
func (h *UserSyncHandler) markDeletedMemberships(instanceID string, seenMembershipKeys map[string]bool, result *UserSyncResult) error {
	// Build list of seen membership keys
	membershipKeys := make([]string, 0, len(seenMembershipKeys))
	for key := range seenMembershipKeys {
		membershipKeys = append(membershipKeys, key)
	}
	
	// For GORM, we need to extract user IDs and group IDs separately
	seenUserIDs := make([]string, 0)
	seenGroupIDs := make([]int, 0)
	
	for key := range seenMembershipKeys {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			seenUserIDs = append(seenUserIDs, parts[0])
			if groupID, err := strconv.Atoi(parts[1]); err == nil {
				seenGroupIDs = append(seenGroupIDs, groupID)
			}
		}
	}
	
	// Mark memberships not in the list as deleted
	now := time.Now()
	res := h.db.Model(&gormmodels.CyberArkGroupMembership{}).
		Where("cyberark_instance_id = ? AND is_deleted = false", instanceID).
		Where("user_id NOT IN ? OR group_id NOT IN ?", seenUserIDs, seenGroupIDs).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		})
	
	if res.Error != nil {
		return res.Error
	}
	
	return nil
}

// processUserVaultAuthorizations processes the vault authorizations for a user
func (h *UserSyncHandler) processUserVaultAuthorizations(instance *gormmodels.CyberArkInstance, caUser *cyberark.User) error {
	userID := strconv.Itoa(caUser.ID)
	
	// Get existing authorizations for this user
	var existingAuths []gormmodels.CyberArkVaultAuthorization
	if err := h.db.Where("cyberark_instance_id = ? AND user_id = ? AND is_deleted = false", 
		instance.ID, userID).Find(&existingAuths).Error; err != nil {
		return fmt.Errorf("query existing authorizations: %w", err)
	}
	
	// Create a map of existing authorizations
	existingAuthMap := make(map[string]*gormmodels.CyberArkVaultAuthorization)
	for i := range existingAuths {
		existingAuthMap[existingAuths[i].Authorization] = &existingAuths[i]
	}
	
	// Track which authorizations we've seen
	seenAuths := make(map[string]bool)
	
	// Process current authorizations
	for _, auth := range caUser.VaultAuthorization {
		seenAuths[auth] = true
		
		if existing, exists := existingAuthMap[auth]; exists {
			// Update existing authorization
			updates := map[string]interface{}{
				"username":       caUser.Username,
				"last_synced_at": time.Now(),
				"is_deleted":     false,
				"deleted_at":     nil,
			}
			
			if err := h.db.Model(existing).Updates(updates).Error; err != nil {
				return fmt.Errorf("update authorization: %w", err)
			}
		} else {
			// Create new authorization
			newAuth := &gormmodels.CyberArkVaultAuthorization{
				CyberArkInstanceID: instance.ID,
				UserID:             userID,
				Username:           caUser.Username,
				Authorization:      auth,
				LastSyncedAt:       time.Now(),
			}
			
			if err := h.db.Create(newAuth).Error; err != nil {
				return fmt.Errorf("create authorization: %w", err)
			}
		}
	}
	
	// Mark authorizations not seen as deleted
	now := time.Now()
	for auth, existing := range existingAuthMap {
		if !seenAuths[auth] {
			updates := map[string]interface{}{
				"is_deleted": true,
				"deleted_at": now,
			}
			
			if err := h.db.Model(existing).Updates(updates).Error; err != nil {
				return fmt.Errorf("mark authorization as deleted: %w", err)
			}
		}
	}
	
	return nil
}

// ValidatePayload validates the operation payload
func (h *UserSyncHandler) ValidatePayload(payload json.RawMessage) error {
	var p UserSyncPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("invalid payload format: %w", err)
	}

	if p.InstanceID == "" {
		return fmt.Errorf("instance_id is required")
	}

	if p.PageSize != nil && *p.PageSize <= 0 {
		return fmt.Errorf("page_size must be greater than 0")
	}

	return nil
}

// Helper function - checks if string contains any of the substrings
func containsSubstring(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(s, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// isRetryableError checks if an error is retryable
func (h *UserSyncHandler) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := err.Error()
	
	// Don't retry client errors (4xx)
	if containsSubstring(errMsg, "400", "401", "403", "404") {
		return false
	}
	
	// Retry server errors and network issues
	return containsSubstring(errMsg, "500", "502", "503", "504", "timeout", "connection", "network")
}

// isAuthError checks if an error is an authentication error
func (h *UserSyncHandler) isAuthError(err error) bool {
	if err == nil {
		return false
	}
	return containsSubstring(err.Error(), "401", "unauthorized", "authentication", "token expired")
}