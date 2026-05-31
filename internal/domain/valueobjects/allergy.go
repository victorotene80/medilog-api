package valueobjects

import "errors"

type AllergyCategory int

const (
	AllergyCategoryFood          AllergyCategory = 1
	AllergyCategoryDrug          AllergyCategory = 2
	AllergyCategoryEnvironmental AllergyCategory = 3
	AllergyCategoryContact       AllergyCategory = 4
	AllergyCategoryOther         AllergyCategory = 5
)

func NewAllergyCategory(v int) (AllergyCategory, error) {
	c := AllergyCategory(v)
	switch c {
	case AllergyCategoryFood, AllergyCategoryDrug,
		AllergyCategoryEnvironmental, AllergyCategoryContact,
		AllergyCategoryOther:
		return c, nil
	}
	return 0, errors.New("invalid allergy category")
}

func (c AllergyCategory) String() string {
	switch c {
	case AllergyCategoryFood:
		return "Food"
	case AllergyCategoryDrug:
		return "Drug"
	case AllergyCategoryEnvironmental:
		return "Environmental"
	case AllergyCategoryContact:
		return "Contact"
	case AllergyCategoryOther:
		return "Other"
	default:
		return "Invalid"
	}
}

func (c AllergyCategory) Int() int { return int(c) }
