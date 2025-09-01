// Package ghbexz provides MCP services for the application.
package ghbexz

import (
	"context"

	"github.com/google/go-github/v61/github"
	ghbex "github.com/rafa-mori/ghbex"
)

// Type aliases from GHbex

type GHbexRules = ghbex.Rules
type GHbexRepoCfg = ghbex.RepoCfg
type GHbexMainConfig = ghbex.MainConfig
type GHbexInsightsReport = ghbex.InsightsReport
type GHbexAutomationReport = ghbex.AutomationReport
type GHbexProductivityReport = ghbex.ProductivityReport
type GHbexRepositoryInsight = ghbex.RepositoryInsight
type GHbexSmartRecommendation = ghbex.SmartRecommendation
type GHbexActivityReport = ghbex.ActivityReport
type GHbexIntelligentSanitizer = ghbex.IntelligentSanitizer
type GHbexSanitizationReport = ghbex.SanitizationReport

// Operators

type IntelligenceOperator = ghbex.IntelligenceOperator
type AutomationService = ghbex.Service

// Bridge functions for Analytics

func AnalyzeRepository(ctx context.Context, client *github.Client, owner, repo string, analysisDays int) (*GHbexInsightsReport, error) {
	return ghbex.AnalyzeRepository(ctx, client, owner, repo, analysisDays)
}

func GetRepositoryInsights(ctx context.Context, owner, repo string, days int) (*GHbexInsightsReport, error) {
	return ghbex.GetRepositoryInsights(ctx, owner, repo, days)
}

// Bridge functions for Automation

func AnalyzeAutomation(ctx context.Context, client *github.Client, owner, repo string, analysisDays int) (*GHbexAutomationReport, error) {
	return ghbex.AnalyzeAutomation(ctx, client, owner, repo, analysisDays)
}

func NewAutomationService(cli *github.Client, cfg GHbexMainConfig) *AutomationService {
	return ghbex.NewService(cli, cfg)
}

// Bridge functions for Productivity

func AnalyzeProductivity(ctx context.Context, client *github.Client, owner, repo string) (*GHbexProductivityReport, error) {
	return ghbex.AnalyzeProductivity(ctx, client, owner, repo)
}

// Bridge functions for Monitoring

func AnalyzeRepositoryActivity(ctx context.Context, cli *github.Client, owner, repo string, inactiveDaysThreshold int) (*GHbexActivityReport, error) {
	return ghbex.AnalyzeRepositoryActivity(ctx, cli, owner, repo, inactiveDaysThreshold)
}

// Bridge functions for Sanitization

func NewIntelligentSanitizer(client *github.Client) *GHbexIntelligentSanitizer {
	return ghbex.NewIntelligentSanitizer(client)
}

// Bridge functions for Intelligence

func NewIntelligenceOperator(cfg GHbexMainConfig, client *github.Client) *IntelligenceOperator {
	return ghbex.NewIntelligenceOperator(cfg, client)
}

// Bridge functions for GitHub Client

func NewGitHubClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}
	return github.NewTokenClient(ctx, token)
}
