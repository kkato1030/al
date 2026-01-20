package profile

import (
	"encoding/json"
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileTemplateCmd creates the profile template command
func NewProfileTemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage profile templates",
		Long:  "Manage profile templates for creating profiles",
	}

	templateCmd.AddCommand(NewProfileTemplateListCmd())
	templateCmd.AddCommand(NewProfileTemplateShowCmd())

	return templateCmd
}

// NewProfileTemplateListCmd creates the template list command
func NewProfileTemplateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		Long:  "List all available profile templates (default and user-defined)",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := config.GetAllTemplates()
			if err != nil {
				return fmt.Errorf("error loading templates: %w", err)
			}

			if len(templates) == 0 {
				fmt.Println("No templates available")
				return nil
			}

			fmt.Println("Available templates:")
			fmt.Println()

			// Get default template names
			defaultTemplates := config.GetDefaultTemplates()
			defaultNames := make(map[string]bool)
			for _, dt := range defaultTemplates {
				defaultNames[dt.Name] = true
			}

			for _, tmpl := range templates {
				prefix := "  "
				if defaultNames[tmpl.Name] {
					prefix = "* "
				}
				fmt.Printf("%s%s", prefix, tmpl.Name)
				if len(tmpl.Profiles) > 0 {
					fmt.Printf(" (creates %d profile(s))", len(tmpl.Profiles))
				}
				fmt.Println()
			}

			fmt.Println()
			fmt.Println("* = default template")

			return nil
		},
	}

	return cmd
}

// NewProfileTemplateShowCmd creates the template show command
func NewProfileTemplateShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show [template-name]",
		Short: "Show template details",
		Long:  "Show detailed information about a profile template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			template, err := config.GetTemplate(templateName)
			if err != nil {
				return fmt.Errorf("error getting template: %w", err)
			}

			if jsonOutput {
				// Output as JSON
				data, err := json.MarshalIndent(template, "", "  ")
				if err != nil {
					return fmt.Errorf("error marshaling template: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Output as human-readable format
			fmt.Printf("Template: %s\n", template.Name)
			fmt.Printf("Profiles to create: %d\n\n", len(template.Profiles))

			for i, profile := range template.Profiles {
				fmt.Printf("Profile %d:\n", i+1)
				fmt.Printf("  Name: %s\n", profile.Name)
				if profile.Stage != "" {
					fmt.Printf("  Stage: %s\n", profile.Stage)
				}
				if profile.Description != "" {
					fmt.Printf("  Description: %s\n", profile.Description)
				}
				if len(profile.Extends) > 0 {
					fmt.Printf("  Extends: %s\n", fmt.Sprintf("%v", profile.Extends))
				}
				if profile.PromoteTo != "" {
					fmt.Printf("  Promote to: %s\n", profile.PromoteTo)
				}
				if profile.PackageDuplication != "" {
					fmt.Printf("  Package duplication: %s\n", profile.PackageDuplication)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}
