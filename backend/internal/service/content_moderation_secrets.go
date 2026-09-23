package service

import (
	"fmt"
	"github.com/AsukaCC/EasySub2api/internal/config"
	"strings"
)

const typeSafeSecretPrefix = "encrypted:v1:"

type contentModerationSecrets struct {
	encryptor  SecretEncryptor
	configured bool
}

func ProvideContentModerationService(settings SettingRepository, repo ContentModerationRepository, hashes ContentModerationHashCache, groups GroupRepository, users UserRepository, proxies ProxyRepository, invalidator APIKeyAuthCacheInvalidator, email *EmailService, encryptor SecretEncryptor, cfg *config.Config) *ContentModerationService {
	return NewContentModerationService(settings, repo, hashes, groups, users, proxies, invalidator, email, contentModerationSecrets{encryptor, cfg.Totp.EncryptionKeyConfigured})
}

func (s *ContentModerationService) encryptTypeSafeConfig(cfg *ContentModerationConfig) (*ContentModerationConfig, error) {
	out := cloneContentModerationConfig(cfg)
	if out.TypeSafe == nil || len(out.TypeSafe.APIKeys) == 0 {
		return out, nil
	}
	if s.secrets.encryptor == nil || !s.secrets.configured {
		return nil, ErrSecretEncryptionKeyNotConfigured
	}
	for i, key := range out.TypeSafe.APIKeys {
		encoded, err := s.secrets.encryptor.Encrypt(key)
		if err != nil {
			return nil, fmt.Errorf("encrypt TypeSafe credentials")
		}
		out.TypeSafe.APIKeys[i] = typeSafeSecretPrefix + encoded
	}
	return out, nil
}

func (s *ContentModerationService) parseStoredModerationConfig(raw string) (*ContentModerationConfig, error) {
	cfg, err := parseContentModerationConfig(raw)
	if err != nil || cfg.TypeSafe == nil {
		return cfg, err
	}
	for i, key := range cfg.TypeSafe.APIKeys {
		if !strings.HasPrefix(key, typeSafeSecretPrefix) {
			continue
		}
		if s.secrets.encryptor == nil {
			return nil, fmt.Errorf("TypeSafe credential encryption unavailable")
		}
		decoded, err := s.secrets.encryptor.Decrypt(strings.TrimPrefix(key, typeSafeSecretPrefix))
		if err != nil {
			return nil, fmt.Errorf("decrypt TypeSafe credentials")
		}
		cfg.TypeSafe.APIKeys[i] = decoded
	}
	return cfg, nil
}
