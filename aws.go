package qsgpm

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mashiike/qsgpm/internal/quicksightx"
)

type QuickSightClient interface {
	ListUsers(context.Context, *quicksight.ListUsersInput, ...func(*quicksight.Options)) (*quicksight.ListUsersOutput, error)
	UpdateUser(ctx context.Context, params *quicksight.UpdateUserInput, optFns ...func(*quicksight.Options)) (*quicksight.UpdateUserOutput, error)

	ListGroups(ctx context.Context, params *quicksight.ListGroupsInput, optFns ...func(*quicksight.Options)) (*quicksight.ListGroupsOutput, error)
	CreateGroup(ctx context.Context, params *quicksight.CreateGroupInput, optFns ...func(*quicksight.Options)) (*quicksight.CreateGroupOutput, error)
	DeleteGroup(ctx context.Context, params *quicksight.DeleteGroupInput, optFns ...func(*quicksight.Options)) (*quicksight.DeleteGroupOutput, error)

	ListGroupMemberships(ctx context.Context, params *quicksight.ListGroupMembershipsInput, optFns ...func(*quicksight.Options)) (*quicksight.ListGroupMembershipsOutput, error)
	CreateGroupMembership(ctx context.Context, params *quicksight.CreateGroupMembershipInput, optFns ...func(*quicksight.Options)) (*quicksight.CreateGroupMembershipOutput, error)
	DeleteGroupMembership(ctx context.Context, params *quicksight.DeleteGroupMembershipInput, optFns ...func(*quicksight.Options)) (*quicksight.DeleteGroupMembershipOutput, error)
}

type QuickSightDryRunClient struct {
	QuickSightClient
}

func (c QuickSightDryRunClient) UpdateUser(ctx context.Context, params *quicksight.UpdateUserInput, optFns ...func(*quicksight.Options)) (*quicksight.UpdateUserOutput, error) {
	bs, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("[notice] **DryRun** UpdateUser input:\n%s\n", string(bs))
	return &quicksight.UpdateUserOutput{
		RequestId: aws.String("<known after run>"),
		Status:    200,
		User: &types.User{
			CustomPermissionsName: params.CustomPermissionsName,
		},
	}, nil
}

func (c QuickSightDryRunClient) CreateGroup(ctx context.Context, params *quicksight.CreateGroupInput, optFns ...func(*quicksight.Options)) (*quicksight.CreateGroupOutput, error) {
	bs, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("[notice] **DryRun** CreateGroup input:\n%s\n", string(bs))
	return &quicksight.CreateGroupOutput{
		RequestId: aws.String("<known after run>"),
		Status:    200,
	}, nil
}

func (c QuickSightDryRunClient) DeleteGroup(ctx context.Context, params *quicksight.DeleteGroupInput, optFns ...func(*quicksight.Options)) (*quicksight.DeleteGroupOutput, error) {
	bs, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("[notice] **DryRun** DeleteGroup input:\n%s\n", string(bs))
	return &quicksight.DeleteGroupOutput{
		RequestId: aws.String("<known after run>"),
		Status:    200,
	}, nil
}

func (c QuickSightDryRunClient) CreateGroupMembership(ctx context.Context, params *quicksight.CreateGroupMembershipInput, optFns ...func(*quicksight.Options)) (*quicksight.CreateGroupMembershipOutput, error) {
	bs, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("[notice] **DryRun** CreateGroupMembership input:\n%s\n", string(bs))
	return &quicksight.CreateGroupMembershipOutput{
		RequestId: aws.String("<known after run>"),
		Status:    200,
	}, nil
}

func (c QuickSightDryRunClient) DeleteGroupMembership(ctx context.Context, params *quicksight.DeleteGroupMembershipInput, optFns ...func(*quicksight.Options)) (*quicksight.DeleteGroupMembershipOutput, error) {
	bs, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("[notice] **DryRun** DeleteGroupMembership input:\n%s\n", string(bs))
	return &quicksight.DeleteGroupMembershipOutput{
		RequestId: aws.String("<known after run>"),
		Status:    200,
	}, nil
}

type QuickSightService struct {
	awsAccountID string
	client       QuickSightClient
}

func getCallerAccountID(ctx context.Context, awsCfg aws.Config) (string, error) {
	client := sts.NewFromConfig(awsCfg)
	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return *output.Account, nil
}

func newQuickSightService(ctx context.Context) (*QuickSightService, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	awsAccountID, err := getCallerAccountID(ctx, awsCfg)
	if err != nil {
		return nil, err
	}
	client := quicksight.NewFromConfig(awsCfg)
	return &QuickSightService{
		awsAccountID: awsAccountID,
		client:       client,
	}, nil
}

func (svc QuickSightService) GetDryRunService() *QuickSightService {
	return &QuickSightService{
		awsAccountID: svc.awsAccountID,
		client: QuickSightDryRunClient{
			QuickSightClient: svc.client,
		},
	}
}

type UsersPaginator struct {
	namespace string
	p         *quicksightx.ListUsersPaginator
}

func (p UsersPaginator) HasMoreUsers() bool {
	return p.p.HasMorePages()
}

func (p UsersPaginator) NextUsers(ctx context.Context) ([]*User, error) {
	output, err := p.p.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*User, 0, len(output.UserList))
	for _, u := range output.UserList {
		users = append(users, &User{
			User:      u,
			Namespace: p.namespace,
		})
	}
	return users, nil
}

func (svc QuickSightService) NewUsersPaginator(namespace string) UsersPaginator {
	p := quicksightx.NewListUsersPaginator(svc.client, &quicksight.ListUsersInput{
		AwsAccountId: aws.String(svc.awsAccountID),
		Namespace:    aws.String(namespace),
	})
	return UsersPaginator{
		namespace: namespace,
		p:         p,
	}
}

func viewStarString(str *string) string {
	if str == nil {
		return "<nil>"
	}
	return *str
}

func (svc QuickSightService) UpdateUserCustomPermission(ctx context.Context, user *User, customPermissionName *string) error {
	log.Printf("[debug] call UpdateUserCustomPermission(ctx, %s, %s)", user, viewStarString(user.CustomPermissionsName))
	if !user.IsNeedUpdateCustomPermission(customPermissionName) {
		log.Printf("[debug] user %s nothing todo", *user.UserName)
		return nil
	}
	input := &quicksight.UpdateUserInput{
		AwsAccountId: aws.String(svc.awsAccountID),
		Namespace:    aws.String(user.Namespace),
		Email:        user.Email,
		UserName:     user.UserName,
		Role:         user.Role,
	}
	if customPermissionName == nil {
		input.UnapplyCustomPermissions = true
	} else {
		input.CustomPermissionsName = customPermissionName
	}
	output, err := svc.client.UpdateUser(ctx, input)
	if err != nil {
		return err
	}
	log.Printf("[info] update user %s custom permission: %s => %s", viewStarString(user.UserName), viewStarString(user.CustomPermissionsName), viewStarString(output.User.CustomPermissionsName))
	return nil
}

func (svc QuickSightService) GetGroups(ctx context.Context, namespace string) (Groups, error) {
	g := newGroups()

	pg := quicksightx.NewListGroupsPaginator(svc.client, &quicksight.ListGroupsInput{
		AwsAccountId: aws.String(svc.awsAccountID),
		Namespace:    aws.String(namespace),
	})
	for pg.HasMorePages() {
		groupsOutput, err := pg.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, group := range groupsOutput.GroupList {
			log.Printf("[debug] group %s exists", *group.GroupName)
			g.AddGroup(*group.GroupName)
			pgm := quicksightx.NewListGroupMembershipsPaginator(svc.client, &quicksight.ListGroupMembershipsInput{
				AwsAccountId: aws.String(svc.awsAccountID),
				Namespace:    aws.String(namespace),
				GroupName:    group.GroupName,
			})
			for pgm.HasMorePages() {
				groupMembershipOutput, err := pgm.NextPage(ctx)
				if err != nil {
					return nil, err
				}
				for _, membership := range groupMembershipOutput.GroupMemberList {
					log.Printf("[debug] group membership %s in %s exists", *membership.MemberName, *group.GroupName)
					g.Add(*group.GroupName, *membership.MemberName)
				}
			}
		}
	}
	return g, nil
}

type ApplyGroupsOptions struct {
	noDeleteGroup           bool
	noDeleteGroupMembership bool
}

func WithCreateOnly(f bool) func(*ApplyGroupsOptions) error {
	return func(opt *ApplyGroupsOptions) error {
		opt.noDeleteGroup = f || opt.noDeleteGroup
		opt.noDeleteGroupMembership = f || opt.noDeleteGroup
		return nil
	}
}

func (svc QuickSightService) ApplyGroups(ctx context.Context, namespace string, groups Groups, optFns ...func(opt *ApplyGroupsOptions) error) error {
	var opts ApplyGroupsOptions
	for _, optFn := range optFns {
		if err := optFn(&opts); err != nil {
			return err
		}
	}

	nowGroups, err := svc.GetGroups(ctx, namespace)
	if err != nil {
		return err
	}
	createGroups, _, deleteGroups := nowGroups.DiffGroup(groups)
	for _, g := range createGroups {
		_, err := svc.client.CreateGroup(ctx, &quicksight.CreateGroupInput{
			AwsAccountId: aws.String(svc.awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(g),
		})
		if err != nil {
			return err
		}
		log.Printf("[info] create group %s", g)
	}
	createMembership, _, deleteMembership := nowGroups.DiffMembership(groups)
	for _, gm := range createMembership {
		_, err := svc.client.CreateGroupMembership(ctx, &quicksight.CreateGroupMembershipInput{
			AwsAccountId: aws.String(svc.awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(gm.GroupName),
			MemberName:   aws.String(gm.UserName),
		})
		if err != nil {
			return err
		}
		log.Printf("[info] create group membership %s in %s", gm.UserName, gm.GroupName)
	}
	if !opts.noDeleteGroupMembership {
		for _, gm := range deleteMembership {
			_, err := svc.client.DeleteGroupMembership(ctx, &quicksight.DeleteGroupMembershipInput{
				AwsAccountId: aws.String(svc.awsAccountID),
				Namespace:    aws.String(namespace),
				GroupName:    aws.String(gm.GroupName),
				MemberName:   aws.String(gm.UserName),
			})
			if err != nil {
				return err
			}
			log.Printf("[info] delete group membership %s in %s", gm.UserName, gm.GroupName)
		}
	}
	if !opts.noDeleteGroup {
		for _, g := range deleteGroups {
			_, err := svc.client.DeleteGroup(ctx, &quicksight.DeleteGroupInput{
				AwsAccountId: aws.String(svc.awsAccountID),
				Namespace:    aws.String(namespace),
				GroupName:    aws.String(g),
			})
			if err != nil {
				return err
			}
			log.Printf("[info] delete group %s", g)
		}
	}
	return nil
}
