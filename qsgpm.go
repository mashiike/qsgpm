package qsgpm

import (
	"context"
	"log"
)

type App struct {
	svc *QuickSightService
	cfg *Config
}

func New(ctx context.Context, cfg *Config) (*App, error) {
	svc, err := newQuickSightService(ctx)
	if err != nil {
		return nil, err
	}
	return &App{
		svc: svc,
		cfg: cfg,
	}, nil
}

type RunOption struct {
	DryRun bool
}

func (app *App) Run(ctx context.Context, opt RunOption) error {
	svc := app.svc
	if opt.DryRun {
		svc = svc.GetDryRunService()
	}
	namespaces := app.cfg.GetNamespaces()
	for _, namespace := range namespaces {
		log.Printf("[debug] namespace: %s", namespace)
		expectGroups := newGroups()
		p := svc.NewUsersPaginator(namespace)
		for p.HasMoreUsers() {
			users, err := p.NextUsers(ctx)
			if err != nil {
				return err
			}
			for _, user := range users {
				if groupNames, ok := app.cfg.GetGroupNames(user); ok {
					expectGroups.Assign(*user.UserName, groupNames)
				}
				customPermissionName := app.cfg.GetCustomPermissionName(user)
				svc.UpdateUserCustomPermission(ctx, user, customPermissionName)
			}
		}
		svc.ApplyGroups(ctx, namespace, expectGroups, WithCreateOnly(app.cfg.CreateOnly))
	}
	return nil
}
