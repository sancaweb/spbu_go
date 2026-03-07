package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
	"strconv"
)

type settingService struct {
	repo repository.SettingRepository
}

func NewSettingService(repo repository.SettingRepository) SettingService {
	return &settingService{repo}
}

func (s *settingService) Get(key string) (string, error) {
	setting, err := s.repo.FindByKey(key)
	if err != nil {
		return "", err
	}
	return setting.SettingValue, nil
}

func (s *settingService) Set(key, value string) error {
	return s.repo.Upsert(key, value)
}

func (s *settingService) GetInt(key string, defaultVal int) int {
	val, err := s.Get(key)
	if err != nil || val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func (s *settingService) GetAll() ([]entity.Setting, error) {
	return s.repo.FindAll()
}
