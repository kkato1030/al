package config

import (
	"testing"
)

func TestStableTrialTemplateHasReviewDays(t *testing.T) {
	templates := GetDefaultTemplates()

	var stableTrialTemplate *ProfileTemplate
	for i := range templates {
		if templates[i].Name == "stable-trial" {
			stableTrialTemplate = &templates[i]
			break
		}
	}

	if stableTrialTemplate == nil {
		t.Fatal("stable-trial template not found")
	}

	// Verify the template has exactly 2 profiles
	if len(stableTrialTemplate.Profiles) != 2 {
		t.Fatalf("Expected 2 profiles in stable-trial template, got %d", len(stableTrialTemplate.Profiles))
	}

	// Find the trial profile (second one)
	var trialProfile *ProfileConfig
	for i := range stableTrialTemplate.Profiles {
		if stableTrialTemplate.Profiles[i].Stage == "trial" {
			trialProfile = &stableTrialTemplate.Profiles[i]
			break
		}
	}

	if trialProfile == nil {
		t.Fatal("trial profile not found in stable-trial template")
	}

	// Verify trial profile has ReviewDays set
	if trialProfile.ReviewDays == nil {
		t.Error("trial profile is missing ReviewDays - this will cause review functionality to not work")
	} else if *trialProfile.ReviewDays != 30 {
		t.Errorf("trial profile ReviewDays expected to be 30, got %d", *trialProfile.ReviewDays)
	} else {
		t.Logf("✓ trial profile has ReviewDays: %d", *trialProfile.ReviewDays)
	}

	// Verify trial profile has PromoteTo set
	if trialProfile.PromoteTo == "" {
		t.Error("trial profile is missing PromoteTo")
	}
}
