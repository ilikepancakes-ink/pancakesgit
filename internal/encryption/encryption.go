package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/ilikepancakes-ink/pancakesgit/internal/config"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// Service provides encryption and decryption functionality
type Service struct {
	key       []byte
	algorithm string
}

// New creates a new encryption service
func New(cfg config.EncryptionConfig) (*Service, error) {
	if cfg.Key == "" {
		return nil, fmt.Errorf("encryption key is required")
	}

	// Derive a 32-byte key from the provided key
	key := deriveKey(cfg.Key)

	return &Service{
		key:       key,
		algorithm: cfg.Algorithm,
	}, nil
}

// deriveKey derives a 32-byte encryption key from the input
func deriveKey(input string) []byte {
	hash := sha256.Sum256([]byte(input))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (s *Service) EncryptString(plaintext string) (string, error) {
	encrypted, err := s.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 encoded string
func (s *Service) DecryptString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	decrypted, err := s.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// HashPassword hashes a password using Argon2id
func (s *Service) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	// Encode salt and hash as base64
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hashB64 := base64.StdEncoding.EncodeToString(hash)
	
	return fmt.Sprintf("$argon2id$%s$%s", saltB64, hashB64), nil
}

// VerifyPassword verifies a password against its hash
func (s *Service) VerifyPassword(password, hashedPassword string) bool {
	// Parse the hash format: $argon2id$salt$hash
	var saltB64, hashB64 string
	if _, err := fmt.Sscanf(hashedPassword, "$argon2id$%s$%s", &saltB64, &hashB64); err != nil {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}

	expectedHash, err := base64.StdEncoding.DecodeString(hashB64)
	if err != nil {
		return false
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	// Constant time comparison
	if len(hash) != len(expectedHash) {
		return false
	}
	
	var result byte
	for i := 0; i < len(hash); i++ {
		result |= hash[i] ^ expectedHash[i]
	}
	
	return result == 0
}

// GenerateToken generates a secure random token
func (s *Service) GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// DeriveKeyFromPassword derives a key from password using scrypt
func (s *Service) DeriveKeyFromPassword(password, salt string) ([]byte, error) {
	saltBytes := []byte(salt)
	key, err := scrypt.Key([]byte(password), saltBytes, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	return key, nil
}

// EncryptWithPassword encrypts data with a password-derived key
func (s *Service) EncryptWithPassword(data []byte, password string) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password
	key, err := s.DeriveKeyFromPassword(password, string(salt))
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Prepend salt and nonce to ciphertext
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptWithPassword decrypts data with a password-derived key
func (s *Service) DecryptWithPassword(data []byte, password string) ([]byte, error) {
	if len(data) < 28 { // 16 (salt) + 12 (nonce) minimum
		return nil, fmt.Errorf("data too short")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:16]
	nonce := data[16:28]
	ciphertext := data[28:]

	// Derive key from password
	key, err := s.DeriveKeyFromPassword(password, string(salt))
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// SecureWipe securely wipes sensitive data from memory
func SecureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
