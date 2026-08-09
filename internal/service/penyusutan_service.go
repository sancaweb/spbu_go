package service

import (
	"strings"
	"time"

	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PenyusutanService interface {
	GenerateFromPenjualan(p *entity.TrxPenjualan, actorID *uint) error
	GetReport(mode, date, month string) ([]repository.PenyusutanReportGroup, error)
	GetShiftReport(date string) ([]repository.PenyusutanShiftGroup, error)
	UpdateEndstockActual(id uint64, actual float64, updatedBy *uint) error
}

type penyusutanService struct {
	repo repository.PenyusutanRepository
}

func NewPenyusutanService(repo repository.PenyusutanRepository) PenyusutanService {
	return &penyusutanService{repo: repo}
}

func (s *penyusutanService) GetShiftReport(date string) ([]repository.PenyusutanShiftGroup, error) {
	return s.repo.GetShiftReport(date)
}

func (s *penyusutanService) GenerateFromPenjualan(p *entity.TrxPenjualan, actorID *uint) error {
	return s.repo.UpsertFromPenjualan(p, actorID)
}

func (s *penyusutanService) GetReport(mode, date, month string) ([]repository.PenyusutanReportGroup, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	date = strings.TrimSpace(date)
	month = strings.TrimSpace(month)

	if mode == "daily" && date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if mode == "monthly" && month == "" {
		month = time.Now().Format("2006-01")
	}

	return s.repo.GetReport(repository.PenyusutanReportFilter{
		Mode:  mode,
		Date:  date,
		Month: month,
	})
}

func (s *penyusutanService) UpdateEndstockActual(id uint64, actual float64, updatedBy *uint) error {
	return s.repo.UpdateEndstockActual(id, actual, updatedBy)
}
