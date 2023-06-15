package quicksightx

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/quicksight"
)

/*
 * The original, original code is here; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/service/quicksight/api_op_ListAnalyses.go#L158
 * The license for the original code is here.; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/LICENSE.txt
 *
 * implemented the ListUsers one by referring to the ListAnalyses paginator.
 * This is a temporary solution.
 */

// ListUsersAPIClient is a client that implements the ListUsers operation.
type ListUsersAPIClient interface {
	ListUsers(context.Context, *quicksight.ListUsersInput, ...func(*quicksight.Options)) (*quicksight.ListUsersOutput, error)
}

// ListUsersPaginatorOptions is the paginator options for ListUsers
type ListUsersPaginatorOptions struct {
	// The maximum number of results to return.
	MaxResults *int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListUsersPaginator is a paginator for ListUsers
type ListUsersPaginator struct {
	options   ListUsersPaginatorOptions
	client    ListUsersAPIClient
	params    *quicksight.ListUsersInput
	nextToken *string
	firstPage bool
}

// NewListUsersPaginator returns a new ListUsersPaginator
func NewListUsersPaginator(client ListUsersAPIClient, params *quicksight.ListUsersInput, optFns ...func(*ListUsersPaginatorOptions)) *ListUsersPaginator {
	if params == nil {
		params = &quicksight.ListUsersInput{}
	}

	options := ListUsersPaginatorOptions{}
	options.MaxResults = params.MaxResults

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListUsersPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListUsersPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListUsers page.
func (p *ListUsersPaginator) NextPage(ctx context.Context, optFns ...func(*quicksight.Options)) (*quicksight.ListUsersOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken
	params.MaxResults = p.options.MaxResults

	result, err := p.client.ListUsers(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}
