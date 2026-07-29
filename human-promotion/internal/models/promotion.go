package models

type Program struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TargetGroup string `json:"target_group"`
	BeneficiariesCount int `json:"beneficiaries_count"`
}
