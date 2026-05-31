package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type AIContext struct {
	Text string
	Meta map[string]any
}

type AIContextBuilder struct {
	profiles    domainRepo.UserProfileRepository
	allergies   domainRepo.UserAllergyRepository
	medications domainRepo.MedicationRepository
	visits      domainRepo.VisitRepository
	drugScans   domainRepo.DrugScanRepository
	clock       func() time.Time
}

func NewAIContextBuilder(
	profiles domainRepo.UserProfileRepository,
	allergies domainRepo.UserAllergyRepository,
	medications domainRepo.MedicationRepository,
	visits domainRepo.VisitRepository,
	drugScans domainRepo.DrugScanRepository,
	clock func() time.Time,
) *AIContextBuilder {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &AIContextBuilder{
		profiles:    profiles,
		allergies:   allergies,
		medications: medications,
		visits:      visits,
		drugScans:   drugScans,
		clock:       clock,
	}
}

func (b *AIContextBuilder) Build(ctx context.Context, userID int64) (*AIContext, error) {
	var out strings.Builder
	meta := map[string]any{
		"user_id": userID,
	}

	out.WriteString("Patient medical context from Medilog records.\n")
	out.WriteString("Use this context only when relevant, and say when the records do not contain enough information.\n")

	if b.profiles != nil {
		profile, err := b.profiles.FindByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("load profile context: %w", err)
		}
		if profile != nil {
			out.WriteString("\nProfile:\n")
			if profile.Weight != nil {
				out.WriteString(fmt.Sprintf("- Weight: %.2f %s\n", *profile.Weight, profile.WeightUnit))
			}
			if profile.Height != nil {
				out.WriteString(fmt.Sprintf("- Height: %.2f\n", *profile.Height))
			}
			out.WriteString(fmt.Sprintf("- AI plan: pro=%t, questions_used=%d, questions_total=%d\n",
				profile.AIIsPro,
				profile.AIQuestionsUsed,
				profile.AIQuestionsTotal,
			))
		}
	}

	if b.allergies != nil {
		allergies, err := b.allergies.FindByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("load allergy context: %w", err)
		}
		ids := make([]string, 0, len(allergies))
		if len(allergies) > 0 {
			out.WriteString("\nKnown allergies:\n")
		}
		for _, allergy := range take(allergies, 12) {
			if allergy == nil {
				continue
			}
			ids = append(ids, allergy.PublicID)
			severity := "not recorded"
			if allergy.Severity != nil {
				severity = allergy.Severity.String()
			}
			out.WriteString(fmt.Sprintf("- %s; category=%s; severity=%s", allergy.Name, allergy.Category.String(), severity))
			if allergy.Description != nil && strings.TrimSpace(*allergy.Description) != "" {
				out.WriteString("; notes=" + strings.TrimSpace(*allergy.Description))
			}
			out.WriteString("\n")
		}
		meta["allergy_public_ids"] = ids
	}

	if b.medications != nil {
		medications, err := b.medications.FindActiveByUserID(ctx, userID, b.clock())
		if err != nil {
			return nil, fmt.Errorf("load medication context: %w", err)
		}
		ids := make([]string, 0, len(medications))
		if len(medications) > 0 {
			out.WriteString("\nActive medications:\n")
		}
		for _, agg := range take(medications, 12) {
			if agg == nil || agg.Medication == nil {
				continue
			}
			m := agg.Medication
			ids = append(ids, m.PublicID)
			out.WriteString(fmt.Sprintf("- %s", m.Name))
			appendKV(&out, "dosage", stringPtrValue(m.Dosage))
			if m.Frequency != nil {
				appendKV(&out, "frequency", m.Frequency.String())
			}
			appendKV(&out, "with_food", fmt.Sprintf("%t", m.WithFood))
			appendKV(&out, "adherence_rate", fmt.Sprintf("%.1f%%", m.AdherenceRate()))
			if len(agg.Times) > 0 {
				times := make([]string, 0, len(agg.Times))
				for _, t := range agg.Times {
					if t != nil {
						times = append(times, t.TimeValue.Format("15:04"))
					}
				}
				appendKV(&out, "times", strings.Join(times, ", "))
			}
			if m.Notes != nil && strings.TrimSpace(*m.Notes) != "" {
				appendKV(&out, "notes", strings.TrimSpace(*m.Notes))
			}
			out.WriteString("\n")
		}
		meta["medication_public_ids"] = ids
	}

	if b.visits != nil {
		visits, err := b.visits.FindByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("load visit context: %w", err)
		}
		ids := make([]string, 0, len(visits))
		if len(visits) > 0 {
			out.WriteString("\nRecent visits:\n")
		}
		for _, visit := range take(visits, 6) {
			if visit == nil {
				continue
			}
			ids = append(ids, visit.PublicID)
			out.WriteString(fmt.Sprintf("- %s", visit.VisitDate.Format("2006-01-02")))
			appendKV(&out, "hospital", stringPtrValue(visit.HospitalName))
			appendKV(&out, "doctor", stringPtrValue(visit.Doctor))
			appendKV(&out, "diagnosis", stringPtrValue(visit.Diagnosis))
			appendKV(&out, "outcome", stringPtrValue(visit.Outcome))
			appendKV(&out, "complaint", stringPtrValue(visit.ChiefComplaint))
			appendKV(&out, "notes", stringPtrValue(visit.Notes))
			if visit.HasVitals() {
				appendVitals(&out, visit.BloodPressure, visit.Temperature, visit.Weight, visit.Pulse)
			}
			out.WriteString("\n")
		}
		meta["visit_public_ids"] = ids
	}

	if b.drugScans != nil {
		scans, err := b.drugScans.FindByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("load drug scan context: %w", err)
		}
		ids := make([]string, 0, len(scans))
		if len(scans) > 0 {
			out.WriteString("\nRecent drug scans:\n")
		}
		for _, scan := range take(scans, 6) {
			if scan == nil {
				continue
			}
			ids = append(ids, scan.PublicID)
			out.WriteString("- Drug scan")
			appendKV(&out, "drug", stringPtrValue(scan.DrugName))
			appendKV(&out, "registration_number", stringPtrValue(scan.RegistrationNumber))
			appendKV(&out, "verified", fmt.Sprintf("%t", scan.IsVerified))
			appendKV(&out, "status", stringPtrValue(scan.VerificationStatus))
			appendKV(&out, "explanation", stringPtrValue(scan.Explanation))
			out.WriteString("\n")
		}
		meta["drug_scan_public_ids"] = ids
	}

	return &AIContext{
		Text: strings.TrimSpace(out.String()),
		Meta: meta,
	}, nil
}

func appendKV(out *strings.Builder, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	out.WriteString(fmt.Sprintf("; %s=%s", key, value))
}

func appendVitals(out *strings.Builder, bp *string, temperature, weight *float64, pulse *int) {
	vitals := make([]string, 0, 4)
	if bp != nil && strings.TrimSpace(*bp) != "" {
		vitals = append(vitals, "bp="+strings.TrimSpace(*bp))
	}
	if temperature != nil {
		vitals = append(vitals, fmt.Sprintf("temperature=%.1f", *temperature))
	}
	if weight != nil {
		vitals = append(vitals, fmt.Sprintf("weight=%.1f", *weight))
	}
	if pulse != nil {
		vitals = append(vitals, fmt.Sprintf("pulse=%d", *pulse))
	}
	if len(vitals) > 0 {
		appendKV(out, "vitals", strings.Join(vitals, ", "))
	}
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func take[T any](items []T, max int) []T {
	if max <= 0 || len(items) <= max {
		return items
	}
	return items[:max]
}
