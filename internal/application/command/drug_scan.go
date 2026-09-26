package command

type VerifyDrugScanCommand struct {
	UserID             int64
	DrugName           *string
	RegistrationNumber *string
	CountryCode        string
}
