package qsgpm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/mashiike/qsgpm"
	"github.com/stretchr/testify/require"
)

func TestConfigValid(t *testing.T) {
	cases := []string{
		"testdata/config.yaml",
		"testdata/config_quicksight.yaml",
	}
	for _, cfgFile := range cases {
		t.Run(cfgFile, func(t *testing.T) {
			cfg := qsgpm.NewDefaultConfig()
			err := cfg.Load(cfgFile)
			require.NoError(t, err)
		})
	}
}

func TestConfigInvalid(t *testing.T) {
	cases := []struct {
		filepath  string
		excpected string
	}{
		{
			filepath:  "testdata/identity_type_invalid.yaml",
			excpected: "rules[0]: user: given IdentityType: Hoge is not one of IAM or QUICKSIGHT",
		},
		{
			filepath:  "testdata/role_invalid.yaml",
			excpected: "rules[1]: user: given Role: Auther is not one of ADMIN, AUTHOR, READER, RESTRICTED_AUTHOR or RESTRICTED_READER",
		},
	}
	for _, c := range cases {
		t.Run(c.filepath, func(t *testing.T) {
			cfg := qsgpm.NewDefaultConfig()
			err := cfg.Load(c.filepath)
			require.EqualError(t, err, c.excpected)
		})
	}
}

func TestConfigGetCustomPermissionName(t *testing.T) {
	cfg := qsgpm.NewDefaultConfig()
	err := cfg.Load("testdata/config.yaml")
	require.NoError(t, err)

	cases := []struct {
		user      *qsgpm.User
		excpected *string
	}{
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("admin@example.com"),
					UserName:     aws.String("Developer/admin@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAdmin,
				},
				Namespace: "default",
			},
			excpected: nil,
		},
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("hoge@example.com"),
					UserName:     aws.String("Manager/hoge@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAuthor,
				},
				Namespace: "default",
			},
			excpected: aws.String("manager"),
		},
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("piyo@example.com"),
					UserName:     aws.String("Analyst/piyo@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAuthor,
				},
				Namespace: "default",
			},
			excpected: aws.String("analysis"),
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			actual := cfg.GetCustomPermissionName(c.user)
			require.Equal(t, c.excpected, actual)
		})
	}
}

func TestConfigGroups(t *testing.T) {
	cfg := qsgpm.NewDefaultConfig()
	err := cfg.Load("testdata/config.yaml")
	require.NoError(t, err)

	cases := []struct {
		user      *qsgpm.User
		excpected []string
	}{
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("admin@example.com"),
					UserName:     aws.String("Developer/admin@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAdmin,
				},
				Namespace: "default",
			},
			excpected: []string{"all", "admins"},
		},
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("hoge@example.com"),
					UserName:     aws.String("Manager/hoge@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAuthor,
				},
				Namespace: "default",
			},
			excpected: []string{"all", "authors", "managers"},
		},
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("piyo@example.com"),
					UserName:     aws.String("Analyst/piyo@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleAuthor,
				},
				Namespace: "default",
			},
			excpected: []string{"all", "authors", "analysts"},
		},
		{
			user: &qsgpm.User{
				User: types.User{
					Email:        aws.String("tora@example.com"),
					UserName:     aws.String("Reader/tora@example.com"),
					IdentityType: types.IdentityTypeIam,
					Role:         types.UserRoleReader,
				},
				Namespace: "default",
			},
			excpected: []string{"all", "readers"},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			actual, ok := cfg.GetGroupNames(c.user)
			require.Equal(t, c.excpected != nil, ok)
			require.ElementsMatch(t, c.excpected, actual)
		})
	}
}
