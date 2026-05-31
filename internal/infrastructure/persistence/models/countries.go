package models

import "github.com/victorotene80/medilog-api/internal/domain/entities"

type CountryModel struct {
	Code     string `gorm:"column:code;primaryKey"`
	Name     string `gorm:"column:name;not null"`
	DialCode string `gorm:"column:dial_code;not null"`
}

func (CountryModel) TableName() string {
	return "countries"
}

func CountryToEntity(m CountryModel) *entities.Country {
	return &entities.Country{
		Code:     m.Code,
		Name:     m.Name,
		DialCode: m.DialCode,
	}
}

func CountryToModel(e entities.Country) *CountryModel {
	return &CountryModel{
		Code:     e.Code,
		Name:     e.Name,
		DialCode: e.DialCode,
	}
}
